package apis

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"template-service-broker/internal"
	"template-service-broker/pkg/server/schemas"

	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
	tmaxv1 "github.com/youngind/hypercloud-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logBind = logf.Log.WithName("binding")

func BindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	var m schemas.ServiceBindingRequest

	//set reponse
	response := &schemas.ServiceBindingResponse{}
	response.Credentials = make(map[string]string)
	w.Header().Set("Content-Type", "application/json")

	//get request body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logBind.Error(err, "error occurs while decoding service binding body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//get url param
	vars := mux.Vars(r)
	serviceId := vars["instance_id"]

	//get templateinstance name & namespace
	instanceName := m.Context["instance_name"] + "." + serviceId
	instanceNameSpace := m.Context["namespace"]

	//add template & template instance schema
	s := scheme.Scheme
	internal.AddKnownTypes(s)
	SchemeBuilder := runtime.NewSchemeBuilder()
	if err := SchemeBuilder.AddToScheme(s); err != nil {
		logBind.Error(err, "cannot add Template/Templateinstance scheme")
	}

	// connect k8s client
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		log.Error(err, "cannot connect k8s api server")
	}

	// get templateinstance info
	templateInstance, err := internal.GetTemplateInstance(c, types.NamespacedName{Name: instanceName, Namespace: instanceNameSpace})
	if err != nil {
		log.Error(err, "cannot get templateinstance info")
	}

	// get template info
	template, err := internal.GetTemplate(c, types.NamespacedName{Name: templateInstance.Spec.Template.Metadata.Name, Namespace: instanceNameSpace})
	if err != nil {
		log.Error(err, "cannot get template info")
	}

	// parse object info in template info
	for _, object := range template.Spec.Objects {
		objectJson, err := json.Marshal(object.Fields)
		if err != nil {
			logBind.Error(err, "error occurs while converting json")
			w.WriteHeader(http.StatusBadRequest)
		}

		//get kind, namespace, name of object
		kind := gjson.Get(string(objectJson), "kind").String()
		namespace := instanceNameSpace
		name := gjson.Get(string(objectJson), "metadata.name").String()
		if gjson.Get(string(objectJson), "metadata.namespace").String() != "" {
			namespace = gjson.Get(string(objectJson), "metadata.namespace").String()
			if strings.Contains(namespace, "{") {
				namespace = getParameter(templateInstance, namespace)
			}
		}
		if name != "" {
			if strings.Contains(name, "{") {
				name = getParameter(templateInstance, name)
			}
		}

		if strings.Compare(kind, "Service") == 0 {
			//set ip in case of service
			service := &corev1.Service{}
			if err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, service); err != nil {
				logBind.Error(err, "error occurs while getting service info")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
				for _, ingress := range service.Status.LoadBalancer.Ingress {
					response.Endpoints.Host = ingress.IP
					response.Credentials["instance-ip"] = ingress.IP
				}
				for _, port := range service.Spec.Ports {
					response.Credentials["instance-port"] = string(strconv.FormatInt(int64(port.Port), 10))
					response.Endpoints.Ports = append(response.Endpoints.Ports, strconv.FormatInt(int64(port.Port), 10))
				}

			}

		} else if strings.Compare(kind, "Secret") == 0 {
			//set credentials in case of secret
			secret := &corev1.Secret{}
			if err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
				logBind.Error(err, "error occurs while getting secret info")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			for key, val := range secret.Data {
				response.Credentials[key] = string(val)
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UnBindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	response := &schemas.AsyncOperation{}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getParameter(ti *tmaxv1.TemplateInstance, s string) string {
	for _, param := range ti.Spec.Template.Parameters {
		if strings.Contains(s, "${"+param.Name+"}") {
			return strings.ReplaceAll(s, "${"+param.Name+"}", param.Value)
		}
	}
	return ""
}
