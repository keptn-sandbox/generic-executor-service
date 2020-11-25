#!/bin/bash

MANIFEST_FILE=./docker/MANIFEST
REPO_URL=https://github.com/$TRAVIS_REPO_SLUG

sed -i 's~MANIFEST_REPOSITORY~'"$REPO_URL"'~' ${MANIFEST_FILE}
sed -i 's~MANIFEST_BRANCH~'"$TRAVIS_BRANCH"'~' ${MANIFEST_FILE}
sed -i 's~MANIFEST_COMMIT~'"$TRAVIS_COMMIT"'~' ${MANIFEST_FILE}
sed -i 's~MANIFEST_TRAVIS_JOB_URL~'"$TRAVIS_JOB_WEB_URL"'~' ${MANIFEST_FILE}
sed -i 's~MANIFEST_DATE~'"$DATE.$TIME"'~' ${MANIFEST_FILE}
sed -i 's~MANIFEST_VERSION~'"$VERSION"'~' ${MANIFEST_FILE}