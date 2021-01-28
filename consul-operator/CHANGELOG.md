# v0.22

* Update operator-sdk version to 0.18.0 from 0.13.0

# v0.21

* Change API group to app.dac.nokia.com from dac.nokia.com 

# v0.8

* Helm timeout increased to 30 sec
* Dynamically request additional resources
* PrivateNetworkAccess support
* Report the IP assigned addresses of the deployment in the appReportedData
* Alarm logging

# v0.7

* DNS Entry example (not used during the deployment, at the moment)

# v0.6

* Ingress support example added

# v0.5

* Fix: in case of the restart of the consul-operator the applied resources will be put again to the CR. This can cause problem during the undeployment. The app can stuck.
* Fix: TweakListOptionsFunc list options was overwritten with a static function, this cause extra events in the informer handler which was not real events 

# v0.4

* Helm3 support
* Logging format changed to json

# v0.3

* Dep changed to go mod
* Operator updated to version 0.13.0

# v0.2

* Memory limit added to the container of the statefulset

# v0.1

* Initial content
