package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	k8sapi "k8s.io/kubernetes/pkg/api"
	unversionedAPI "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"
    "github.com/frankgreco/kubenforce/spec"
    "github.com/frankgreco/kubenforce/utils"
)

const (
	tprName = "config-policy.k8s.io"
)

var (
	ErrVersionOutdated = errors.New("Requested version is outdated in apiserver")
	initRetryWaitTime  = 30 * time.Second
)

type Controller struct {
	logger         *logrus.Entry
	Config         Config
	waitFunction   sync.WaitGroup
	ConfigPolicies map[string]*spec.ConfigPolicy
}

type Config struct {
	Namespace  string
	KubeCli    *unversioned.Client
	MasterHost string
}

type rawEvent struct {
	Type   string
	Object json.RawMessage
}

type Event struct {
	Type   string
	Object *spec.ConfigPolicy
}

func New(cfg Config) *Controller {
    return &Controller{
		logger:           logrus.WithField("pkg", "controller"),
		Config:           cfg,
		ConfigPolicies:   make(map[string]*spec.ConfigPolicy),
	}
}

func (c *Controller) Init() {
    c.logger.Infof("Init started!")
	for {
		//create TPR if it's not exists
		err := c.initResource()
		if err == nil {
			break
		}
		c.logger.Errorf("Initialization failed: %v", err)
		c.logger.Infof("Retry in %v...", initRetryWaitTime)
		time.Sleep(initRetryWaitTime)
	}
}

func (c *Controller) initResource() error {
	if c.Config.MasterHost == "" {
		return fmt.Errorf("MasterHost is empty. Please check if k8s cluster is available.")
	}
	err := c.createTPR()
	if err != nil {
		if !utils.IsKubernetesResourceAlreadyExistError(err) {
			return fmt.Errorf("Fail to create TPR: %v", err)
		}
	}
    c.logger.Infof("TPC created!")
	return nil
}

func (c *Controller) createTPR() error {
	tpr := &extensions.ThirdPartyResource{
		ObjectMeta: k8sapi.ObjectMeta{
			Name: tprName,
		},
		Versions: []extensions.APIVersion{
			{Name: "v1"},
		},
		Description: "my description",
	}
    c.logger.Infof("about to try create the resource")
	_, err := c.Config.KubeCli.ThirdPartyResources().Create(tpr)
	if err != nil {
        c.logger.Infof("initial error")
		return err
	}
    c.logger.Infof("TPC created!")
	return nil
}

func (c *Controller) Run() error {
	defer func() {
		c.waitFunction.Wait()
	}()

	eventCh, errCh := c.monitor()

	go func() {
		for event := range eventCh {
			switch event.Type {
			case "ADDED":
				c.logger.Infof("a new config policy was added: %s", event.Object.ObjectMeta.Name)
                event.Object.RetroFit(c.Config.KubeCli)

            }
		}
	}()
	return <-errCh
}

func (c *Controller) monitor() (<-chan *Event, <-chan error) {
	host := c.Config.MasterHost
	ns := c.Config.Namespace
	httpClient := c.Config.KubeCli.RESTClient.Client

	eventCh := make(chan *Event)
	// On unexpected error case, controller should exit
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		for {
			resp, err := utils.WatchResources(host, ns, httpClient)
			if err != nil {
				errCh <- err
				return
			}
			if resp.StatusCode != 200 {
				resp.Body.Close()
				errCh <- errors.New("Invalid status code: " + resp.Status)
				return
			}
			decoder := json.NewDecoder(resp.Body)
			for {
				ev, _, err := pollEvent(decoder)

				if err != nil {
					if err == io.EOF { // apiserver will close stream periodically
						c.logger.Debug("Apiserver closed stream")
						break
					}

					c.logger.Errorf("Received invalid event from API server: %v", err)
					errCh <- err
					return
				}

				eventCh <- ev
			}

			resp.Body.Close()
		}
	}()

	return eventCh, errCh
}

func pollEvent(decoder *json.Decoder) (*Event, *unversionedAPI.Status, error) {
	re := &rawEvent{}
	err := decoder.Decode(re)
	if err != nil {
		if err == io.EOF {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("Fail to decode raw event from apiserver (%v)", err)
	}

	if re.Type == "ERROR" {
		status := &unversionedAPI.Status{}
		err = json.Unmarshal(re.Object, status)
		if err != nil {
			return nil, nil, fmt.Errorf("Fail to decode (%s) into unversioned.Status (%v)", re.Object, err)
		}
		return nil, status, nil
	}

	ev := &Event{
		Type:   re.Type,
		Object: &spec.ConfigPolicy{},
	}
	err = json.Unmarshal(re.Object, ev.Object)
	if err != nil {
		return nil, nil, fmt.Errorf("Fail to unmarshal function object from data (%s): %v", re.Object, err)
	}
	return ev, nil, nil
}
