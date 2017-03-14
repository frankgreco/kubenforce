#!/bin/bash

ROOT=$PWD

for SKILL in skills/*/ ; do
    cd $SKILL/src
    FUNCTION_NAME=$(cat package.json | jq -r '.name')
    ARN=arn:aws:lambda:$AWS_DEFAULT_REGION:$ACCOUNT_NUMBER:function:$FUNCTION_NAME
    npm install
    zip -r $FUNCTION_NAME.zip *
    aws lambda update-function-code --function-name $ARN --zip-file fileb://$FUNCTION_NAME.zip --publish
    cd $SKILLS_DIR
done
