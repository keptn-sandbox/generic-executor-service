# Sample files

This directory contains a couple of sample files that I am using for my own testing. 
Feel free to use them as well by adding them to your Keptn Project for a specific service in a specific stage. You can either add this via the Keptn CLI, the API or just by adding it to your Upstream Git Repo. Here an example Keptn CLI call to add these scripts to a specific project:

```
keptn add-resource --project=PROJECTNAME --stage=STAGE --service=SERVICENAME --resource=./all.events.sh --resourceUri=generic-executor/all.events.sh
keptn add-resource --project=PROJECTNAME --stage=STAGE --service=SERVICENAME --resource=./configuration.change.http --resourceUri=generic-executor/configuration.change.http
```

Also remember: if you want to have these scripts or http requests executed for certain Keptn events simply give it the corresponding event name. Here is an example to use one of my sample scripts just for the tests.finished event:
```
keptn add-resource --project=PROJECTNAME --stage=STAGE --service=SERVICENAME --resource=./all.events.sh --resourceUri=generic-executor/tests.finished.sh
```

To get an overview of all possible events and the corresponding file names please check out the [Generic Executor Service Readme](../readme.md)