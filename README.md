# ce-event-passthrough
A simple KNative Service to be used within KNative Flows for event investigation. 

Heavily inspired by [KNative Event Display](https://github.com/knative/eventing/blob/main/cmd/event_display/main.go).

# Usage - Kubernetes

You can deploy this service as a `ksvc` into your cluster and inject it into your `flows` (i.e. `Sequences`). Please refer to `deployments/ce-event-passthrough.yaml` for sample Service definition.


# Fetch Events

The events are logged in string format to the stdout - you can use `kubectl logs` command to fetch the logs from your service. Service exposes also an endpoint `/eventz` which you can `curl` to fetch the last 100 events that went through the service. 

- Example: `curl {SERVICE_ADDR}/eventz | jq`

```
[
  {
     "specversion":"1.0",
     "id":"59cbc490-1722-4a40-9039-15c1127719c4",
     "source":"/apis/v1/namespaces/development/mqttsources/mqttsource#sample-manufacturingdata",
     "type":"dev.service.source-service.event",
     "datacontenttype":"application/json",
     "time":"2022-11-04T13:59:04.361634Z",
     "data":{
        "device":"sample_device",
        "value":123
     }
  }
]
```