# Application operator template/example for the NDAC Application Framework

The purpose of this project is to show an example how a simple application operator can be created which is fully
compatible with the NDAC Application Framework (NDAC App FW). 

It shows the way how the NDAC App FW platform services can be requested, how a simple application can be 
deployed/undeployed and how an application can send back data to the DC side (appStatus, appReportedData). 
It provides these functionalities as a general solution which can be easily reused by most of the app operators.

In this specific example the application is [Consul](https://www.consul.io/) which is a well known key-values data store.
It is a stateful application so it needs persistent storage and it has prometheus endpoint to provide the metrics.
That's why it is a good candidate to represent the currently available NDAC App FW platform services.

## Features
#### Platform resource request handling
The only way for an application to request NDAC platform resources (persistent storage, prometheus monitoring collection,
access to the PLTE network, etc.) is to create the CR (CustomResource) of the platform resource provider. A CR contains
the parameters which are needed to give the requested resource for the application. Every platform resource provider
has its own CR. Based on the given parameters in the CR and the available application constraints the platform resource
provider evaluates the request. If the requested resource can be given for the application then the provider does the 
needed configuration and update the result of the evaluation in the CR (status/approvalStatus: Approved or Rejected).

This example contains a metrics collection and a storage request. The application deployment starts with the apply of 
these requests and the deployment flow continuous only when the resources are granted for the application.

#### Ingress for the application Components
This project has an example how the application components which have HTTP interface can be reachable from outside,
using a domain name. The domain name should come from the app spec CR, defined by the customer. The customer needs to
register the domain name in his/her own DNS server which should point to the ingress controller of the application
framework.
Applications should use the domain name(s) given in the app spec CR and create [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) 
resources, like this [example]((deployment/app-deployment/templates/consul-ingress.yaml))
To test if you can reach the service of the application via the ingress controller, you can use a PC which can access the
IP of the ingress controller, and execute the follwing command:
```
curl --noproxy metrics.consul.appdomain.com http://10.10.38.3/v1/agent/metrics -H 'Host: metrics.consul.appdomain.com'
```
In this example the 10.10.38.3 address is the IP of the ingress controller. The entry is not registred in the DNS server or
in the hosts file of the machine so it is defined as an additional Host header.
 
#### CR based deployment templating
An application operator has at least one CR (currently exactly one supported by the App FW) which contains the basic
configuration of the application and serves as a trigger for the deployment/undeplyoment.
The values from the CR can be used as template variables which can be inserted into the deployment yamls.
The yaml files of your application should be located under the deployment/app-deployment directory.

Example CR:
```yaml
apiVersion: app.dac.nokia.com/v1alpha1
kind: Consul
metadata:
  name: example-consul
spec:
  replicaCount: 1
  ports:
    uiPort: 8500
    altPort: 8400
    udpPort: 53
    httpPort: 8080
    httpsPort: 8443
    serflan: 8301
    serfwan: 8302
    consulDns: 8600
    server: 8300
```

Example usage in the deployment yaml.
```yaml
replicaCount: [[ .ReplicaCount ]]

service:
  uiport: [[ .Ports.UiPort ]]
  altport: [[ .Ports.AltPort ]]
  udpport: [[ .Ports.UdpPort ]]
  httpsport: [[ .Ports.HttpPort ]]
  httpport: [[ .Ports.HttpsPort ]]
  serflan: [[ .Ports.Serflan ]]
  serfwan: [[ .Ports.Serfwan ]]
  consuldns: [[ .Ports.ConsulDns ]]
  server: [[ .Ports.Server ]]
```
The name of the variable comes from the defined go structure [consul_types.go](pkg/apis/dac/v1alpha1/consul_types.go)
```go
type ConsulSpec struct {
	ReplicaCount int   `json:"replicaCount"`
	Ports        Ports `json:"ports"`
}

type Ports struct {
	UiPort    int `json:"uiPort,omitempty"`
	AltPort   int `json:"altPort,omitempty"`
	UdpPort   int `json:"udpPort,omitempty"`
	HttpPort  int `json:"httpPort,omitempty"`
	HttpsPort int `json:"httpsPort,omitempty"`
	Serflan   int `json:"serflan,omitempty"`
	Serfwan   int `json:"serfwan,omitempty"`
	ConsulDns int `json:"consulDns,omitempty"`
	Server    int `json:"server,omitempty"`
}
```

To execute this CR based templating you need to create a templater object and call the RunCrTemplater method. It gives
back the templated, concatenated yaml files as a string which can be applied in the Kubernetes:
```go
	templater, err := template.NewTemplater(instance.Spec, namespace)
	if err != nil {
		logger.Error(err, "Failed to initialize the templater")
		return reconcile.Result{}, nil
	}

	out, err = templater.RunCrTemplater("---\n")
	if err != nil {
		logger.Error(err, "Failed to execute the CR templater")
		return reconcile.Result{}, nil
	}
```

#### Helm3 chart support
If your application already has a Helm chart then you can reuse that in the application operator. This example 
project uses Helm3 to deploy/undeploy the chart placed in the app-deployment directory.
If you use the CR templating, described in the previous section, then you can feed the data from the CR to the values.yaml
of your chart or to any other yaml in the app-deployment directory.

Usage:
```go
	//Execute templating for the app-deplyoment directory using the values from the CR
	templater, err := template.NewTemplater(instance.Spec, namespace)
	if err != nil {
		logger.Error(err, "Failed to initialize the templater")
		return reconcile.Result{}, nil
	}

	//Gives back the output of the templated yamls. It can be applied directly in kubernetes
	_, err = templater.RunCrTemplater("---\n")
	if err != nil {
		logger.Error(err, "Failed to execute the CR templater")
		return reconcile.Result{}, nil
	}

	//Optional - Helm based deployment
	err = helm.NewHelm(namespace).Deploy()
	if err != nil {
		logger.Error(err, "Failed to deploy the helm chart")
		return reconcile.Result{}, nil
	}
```

The helm chart support and the CR based templating is independent from each other. You can use one or both of them.
In this example the parameters in the values.yaml are filled from the CR using the CR templating feature.
 
#### Application pod status monitoring
An application operator must report back the status of the application via its own CR in the status/appStatus field.
This information will be used on the NDAC customer portal to show whether the application is working or not.

This example has a very basic monitoring capability. It checks the status of those pods which have the `statusCheck: "true"`
label in the pod definition. If all of the pods are in `Running` state which has the mentioned label then the appStatus
is updated to `Running`. When the status of one of the pod changes from `Running` to anything else, the 
appStatus is updated to `NotRunning`.

A complex application needs to have a more sophisticated mechanism to handle this status update but this is 
absolutely application specific.

#### Reporting data to NDAC
An application operator has the possibility to report back some custom data to the NDAC DC. This data can be
visualized on the NDAC Customer or Maintenance UI. Typically such data should be reported which gets value after
the deployment and its value can not be known before. Eg: the url of the application UI or some dynamic IP address
of an application component. The structure of the reported data is up to the application. It can be a complex
or a simple data structure since it is transformed to a JSON representation. Anything under the status/appReportedData
structure in the CR of the application operator is converted to JSON and transferred to the DB of the NDAC DC
in case of every change. The CRD (CustomResourceDefinition) of the application operator should contain the
OpenApi schema of the appReportedData. In this way, the UI can generate an application specific page to
visualize the reported information.

This feature of the example depends on the monitoring feature. A running and a notRunning callback can 
be defined for the Monitoring service which will be called when the status of the application changes.
App FW watches all of the changes of the appReportedData and sends it to the DC. 

Example:
```go
	monitor := monitoring.NewMonitor(r.client, instance, namespace,
		func() {
			logger.Info("Set AppReportedData")
			//runningCallback - example, some dynamic data should be reported here which has value only after the deployment
			svc, err := kubelib.GetKubeAPI().CoreV1().Services(namespace).Get("consul-operator-metrics", v1.GetOptions{})
			if err != nil {
				logger.Error(err, "Failed to read the svc of the metrics endpoint")
				return
			}
			instance.Status.AppReportedData.MetricsClusterIp = svc.Spec.ClusterIP
			if err := r.client.Status().Update(context.TODO(), instance); nil != err {
				logger.Error(err, "status app reported data update failed")
			}
		},
		func() {
			//notRunningCallback
		},
	)
	monitor.Run()
```

#### Application licence handling
Every application in NDAC App FW has a corresponding licence to protect it from unwanted use. It is the
the responsibility of the application developer to define its behavior in the event of licence expiration
and re/activation, thus the developer should implement the following interface:
```go
type LicenceExpiredResourceFuncs interface {
	Expired()
	Activate()
}
```

The implementation (instance) of above interface can then be passed as the second argument in instantiating
licenceexpired Handler:
```go
func New(namespace string, callbacks LicenceExpiredResourceFuncs) *Handler
```
Example:
```go
licenceexpired.New(namespace, licCallbacks).Watch()
```
where licCallbacks implements LicenceExpiredResourceFuncs interface.

It is recommended to set the application status appropriately. When its licence expired, the operator
should be able to set its status to "FROZEN" so the NDAC App FW will be notified. Also, when its licence
becomes valid, its status should be set to "RUNNING" or "NOT_RUNNING" depending on the application
criteria set for it to be in either state.

#### Application removal
In case the CR of the application operator is deleted the operator should gracefully stop the application
and removed the deployed resources. It again depends on the application how it can be safely stopped.

This example project stores the GVK (Group, Version, Kind) of every resource which was applied in the 
Kubernetes and in case of a CR delete it starts to delete every resource using this information. When it 
is finished, it removes the finalizer from the CR to indicate to the App FW that the application 
is terminated and its namespace can be deleted.  

## Steps to create your own application operator
Prerequirement:  [operator-sdk](https://github.com/operator-framework/operator-sdk) cli is needed for the
following commands. 

1. First you should add a new API resource which will be managed by your application operator. The example
project has the Consul resource under the pkg/apis/dac/v1alpha1 directory. You can choose to rename this
according to your need but probably it is easier to generate a new one and copy the relevant parts from
the Consul API. After that you can remove the Consul API.
It is mandatory to include the appStatus field in the status section and it is suggested to put the 
appReportedData as well:
  
    ```go
    const (
        AppStatusNotSet     = "UNSET"
        AppStatusNotRunning = "NOT_RUNNING"
        AppStatusRunning    = "RUNNING"
        AppStatusFrozen     = "FROZEN"
    )
    
    type AppReporteData struct {
    	//this stucture can be anything
    }
    
    // ConsulStatus defines the observed state of Consul
    // +k8s:openapi-gen=true
    type ConsulStatus struct {
    	AppStatus        AppStatus                       `json:"appStatus,omitempty"`
    	AppReportedData  AppReporteData                  `json:"appReportedData,omitempty"`
    }
    ```
   
    Command to generate your own API resource:
    >operator-sdk add api --api-version=app.dac.nokia.com/v1alpha1 --kind=Consul
  
2. Generate the controller for your new API resource. You can again reuse the controller of the Consul
resource or you can generate your own and copy the useful parts from the Consul controller (pkg/controller/consul).

    Command to generate the controller of your own resource:
    >operator-sdk add controller --api-version=app.dac.nokia.com/v1alpha1 --kind=Consul

3. Replace the content of the deployment/app-deployment and deployment/resource-reqs directories with your
custom application yamls.

4. For the application lifecycle management, App FW uses the [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager)
components. It has an application registry which should contain the metadata of each and every version
of an application. In practice it has information only about the application specific operator. To create
this registry you need to generate a CSV (ClusterServiceVersion) file for you operator.

    Command to generate the skeleton of the CSV (the result is generated under deploy/olm-catalog directory):
    >operator-sdk olm-catalog gen-csv --csv-version 0.0.1 --operator-name consul

    Note: from version v0.15.0 of the operator-sdk the `olm-catalog gen-csv` subcommand was moved to `generate csv`
    So the command above should look like this:
    >operator-sdk generate csv --csv-version 0.0.1 --operator-name consul                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       >                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   >
 
    You need to insert an exact docker image of your application specific operator to the CSV  file.
 
    It can be built with the following command:
    > operator-sdk build docker-registry.vepro.nsn-rdnet.com/appfw/consul-operator:0.1
 
    After that you should replace the `REPLACE_IMAGE` placeholder in the CSV with your docker image name. 
    
    Pull secret should be added to the CSV file. During testing it should be the pull secret of your own registry.
    
    !!! IMPORTANT !!! this secret should be part of the application registry, so it should be next to the CSV file,
    but you also need to create the same secret in the `olm` namespace. OLM will create it automatically in
    the application namespace if it can find the same secret also in the `olm` namespace. 
     
    ```yaml
                  imagePullSecrets:
                    - name: pull-secret
    ```

5. Application constraints should be added to the CSV file. This a descriptor to an exact version of the application
about the needed platform resources. When your application will be part of the official application registry
these resource needs will be checked and approved by the NDAC team. After that the platform service provider
components won't let your application to request and access more resources than you have in the constraints section.                
Currently you should describe only your storage needs.

    Example application constraint:
    ```yaml
    apiVersion: operators.coreos.com/v1alpha1
    kind: ClusterServiceVersion
    metadata:
      name: indoor.v0.0.1
      namespace: placeholder
      annotations: 
        application-constraints: |
          storage-service: 
            requests:
            - name: storage-for-db
              size: 500Mi   
    spec:
    ```
6. The `customresourcedefinitions` section in the CSV file should be extended with the `displaName` and `description`
fields.
    ```yaml
      customresourcedefinitions:
        owned:
        - kind: Consul
          name: consuls.app.dac.nokia.com
          version: v1alpha1
    +     displayName: Consul application
    +     description: Consul application
    ```
   
7. The binary name of your application operator should be check and updated in the CSV file.
    ```yaml
                spec:
                  containers:
                  - command:
                    - consul-operator
    ```
8. To access the platform resources the RBAC section should be extended in the CSV file with the API group of the 
platform request resources:
    ```yaml
            - apiGroups:
                - ops.dac.nokia.com
              resources:
                - '*'
                - consuls
              verbs:
                - '*'
            serviceAccountName: consul-operator
    ```                  

## How to test your application operator without App FW in your local environment
1. After your operator will be able to build. You will need a Kubernetes cluster and there
you need to deploy your operator by applying the yaml files from the deploy directory. 

    What you need are the:
    -	deploy/role.yaml
    -	deploy/rolebindig.yaml
    -	deploy/service_account.yaml
    -	deploy/crd/app.dac.nokia.com_consuls_crd.yaml
    -	deploy/operator.yaml (this part need to be replaced “image: REPLACE_IMAGE”)

2. The next step is to apply the CR from the deploy/crds/*_cr.yaml to the same namespace where your
operator is running. For this phase you should delete the content of the deployment/resource-reqs directory because on your
environment the NDAC platform resource providers are not available so your operator won’t be able to get the needed resources
and it will interrupt the deployment.

    With any empty resource-reqs directory you will see that your operator is deploying your application and when you delete
    the CR it should delete the deployed components.

3. After this phase works well you can proceed with the OLM integration.
You should install the OLM components in your k8s cluster. It can be done by executing the install.sh from here :  
https://github.com/operator-framework/operator-lifecycle-manager/tree/0.13.0/deploy/upstream/quickstart 

    You will have an olm namespace with some components. The olm-operator is the only relevant for you.
    You need to create a new namespace where your operator will be deployed. And in that namespace first you should create an OLM
    specific OperatorGroup resource. 

    This is an example, please replace the “your-namespace” string with the namespace where you want to install your operator.

    ```yaml
    apiVersion: operators.coreos.com/v1
    kind: OperatorGroup
    metadata:
      name: example-operatorgroup
      namespace: your-namespace
    spec:
      targetNamespaces:
      - your-namespace
    ```

4. You should apply your CRD and CSV file(deploy/olm-catalog/consul/0.0.1/consul.v0.0.1.clusterserviceversion.yaml) in the
newly created namespace. The CSV will be detected by the olm-operator and it will deploy the operator. If there is
some problem, you can check the status part of the installPlan and CSV resources. It should contain the reason why OLM can’t do
the deployment.

5. The last step is the CR creation to trigger the application deployment.

    
## How to test your application operator with the App FW
1. First you need to build an [application-registry](https://github.com/operator-framework/operator-registry) which 
contains your new application. 

    You can build it manually by cloning [this](https://github.com/operator-framework/operator-registry)
    repo. After that you should replace the content of the manifest directory with the deploy/olm-catalog. And finally
    you should build and push the docker image with these commands:
    
    >docker build -t {replace this with your image name} -f upstream-example.Dockerfile .
    >docker push {replace this with your image name}       
    
    After you have the new application registry image with your own content in the docker registry, you just need to 
    update the app-registry deployment in your edge which is running in the `appfw` namespace.

2. At this phase you are still not able to test the application deployment in e2e because the DC side components also
need to know about your new application.(app-catalog should be extended with the new app, new license should be 
created for the application and finally the this new license should be activated for the user). So you should request
these modifications from the NDAC team.

3. After the previously mentioned requirements are fulfilled you can trigger CR upload, Deployment, Undeployment
operations via REST API.





























     

