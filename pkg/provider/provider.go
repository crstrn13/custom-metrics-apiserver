package provider

import (
	"context"
	"net/http"
	"sync"

	"github.com/emicklei/go-restful/v3"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/defaults"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/helpers"
)

// CustomMetricResource wraps provider.CustomMetricInfo in a struct which stores the Name and Namespace of the resource
// So that we can accurately store and retrieve the metric as if this were an actual metrics server.
type CustomMetricResource struct {
	provider.CustomMetricInfo
	types.NamespacedName
}

type metricValue struct {
	labels    labels.Set
	value     resource.Quantity
	timestamp metav1.Time
}

// testingProvider is a sample implementation of provider.MetricsProvider which stores a map of fake metrics
type testingProvider struct {
	defaults.DefaultCustomMetricsProvider
	defaults.DefaultExternalMetricsProvider
	client dynamic.Interface
	mapper apimeta.RESTMapper

	valuesLock sync.RWMutex
	values     map[CustomMetricResource]metricValue
}

// NewFakeProvider creates a new testingProvider and returns it along with a restful.WebService
func NewFakeProvider(client dynamic.Interface, mapper apimeta.RESTMapper) (provider.CustomMetricsProvider, *restful.WebService) {
	provider := &testingProvider{
		client: client,
		mapper: mapper,
		values: make(map[CustomMetricResource]metricValue),
	}
	return provider, provider.webService()
}

// webService creates a restful.WebService with routes set up for receiving fake metrics
func (p *testingProvider) webService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/write-metrics")

	// Namespaced resources
	ws.Route(ws.POST("/namespaces/{namespace}/{resourceType}/{name}/{metric}").To(p.updateMetric).
		Param(ws.BodyParameter("value", "value to set metric").DataType("integer").DefaultValue("0")))

	// Root-scoped resources
	ws.Route(ws.POST("/{resourceType}/{name}/{metric}").To(p.updateMetric).
		Param(ws.BodyParameter("value", "value to set metric").DataType("integer").DefaultValue("0")))

	// Namespaces, where {resourceType} == "namespaces" to match API
	ws.Route(ws.POST("/{resourceType}/{name}/metrics/{metric}").To(p.updateMetric).
		Param(ws.BodyParameter("value", "value to set metric").DataType("integer").DefaultValue("0")))
	return ws
}

// updateMetric writes the metric provided by a restful request and stores it in memory
func (p *testingProvider) updateMetric(request *restful.Request, response *restful.Response) {
	p.valuesLock.Lock()
	defer p.valuesLock.Unlock()

	namespace := request.PathParameter("namespace")
	resourceType := request.PathParameter("resourceType")
	namespaced := len(namespace) > 0 || resourceType == "namespaces"

	name := request.PathParameter("name")
	metricName := request.PathParameter("metric")

	value := new(resource.Quantity)
	err := request.ReadEntity(value)
	if err != nil {
		if err := response.WriteErrorString(http.StatusBadRequest, err.Error()); err != nil {
			klog.Errorf("Error writing error: %s", err)
		}
		return
	}

	groupResource := schema.ParseGroupResource(resourceType)

	metricLabels := labels.Set{}
	sel := request.QueryParameter("labels")
	if len(sel) > 0 {
		metricLabels, err = labels.ConvertSelectorToLabelsMap(sel)
		if err != nil {
			if err := response.WriteErrorString(http.StatusBadRequest, err.Error()); err != nil {
				klog.Errorf("Error writing error: %s", err)
			}
			return
		}
	}

	info := provider.CustomMetricInfo{
		GroupResource: groupResource,
		Metric:        metricName,
		Namespaced:    namespaced,
	}

	info, _, err = info.Normalized(p.mapper)
	if err != nil {
		klog.Errorf("Error normalizing info: %s", err)
	}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	metricInfo := CustomMetricResource{
		CustomMetricInfo: info,
		NamespacedName:   namespacedName,
	}
	p.values[metricInfo] = metricValue{
		labels:    metricLabels,
		value:     *value,
		timestamp: metav1.Now(),
	}
}

// valueFor is a helper function to get just the value of a specific metric
func (p *testingProvider) valueFor(info provider.CustomMetricInfo, name types.NamespacedName, metricSelector labels.Selector) (metricValue, error) {
	info, _, err := info.Normalized(p.mapper)
	if err != nil {
		return metricValue{}, err
	}
	metricInfo := CustomMetricResource{
		CustomMetricInfo: info,
		NamespacedName:   name,
	}

	value, found := p.values[metricInfo]
	if !found {
		return metricValue{}, provider.NewMetricNotFoundForError(info.GroupResource, info.Metric, name.Name)
	}

	if !metricSelector.Matches(value.labels) {
		return metricValue{}, provider.NewMetricNotFoundForSelectorError(info.GroupResource, info.Metric, name.Name, metricSelector)
	}

	return value, nil
}

// metricFor is a helper function which formats a value, metric, and object info into a MetricValue which can be returned by the metrics API
func (p *testingProvider) metricFor(value metricValue, name types.NamespacedName, _ labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	objRef, err := helpers.ReferenceFor(p.mapper, name, info)
	if err != nil {
		return nil, err
	}

	metric := &custom_metrics.MetricValue{
		DescribedObject: objRef,
		Metric: custom_metrics.MetricIdentifier{
			Name: info.Metric,
		},
		Timestamp: value.timestamp,
		Value:     value.value,
	}

	if len(metricSelector.String()) > 0 {
		sel, err := metav1.ParseToLabelSelector(metricSelector.String())
		if err != nil {
			return nil, err
		}
		metric.Metric.Selector = sel
	}

	return metric, nil
}

// metricsFor is a wrapper used by GetMetricBySelector to format several metrics which match a resource selector
func (p *testingProvider) metricsFor(namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	names, err := helpers.ListObjectNames(p.mapper, p.client, namespace, selector, info)
	if err != nil {
		return nil, err
	}

	res := make([]custom_metrics.MetricValue, 0, len(names))
	for _, name := range names {
		namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
		value, err := p.valueFor(info, namespacedName, metricSelector)
		if err != nil {
			if apierr.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		metric, err := p.metricFor(value, namespacedName, selector, info, metricSelector)
		if err != nil {
			return nil, err
		}
		res = append(res, *metric)
	}

	return &custom_metrics.MetricValueList{
		Items: res,
	}, nil
}

// GetMetricByName is a wrapper used by GetMetricByName to format a metric which matches a resource name
func (p *testingProvider) GetMetricByName(_ context.Context, name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	value, err := p.valueFor(info, name, metricSelector)
	if err != nil {
		return nil, err
	}
	return p.metricFor(value, name, labels.Everything(), info, metricSelector)
}

// GetMetricBySelector is a wrapper used by GetMetricBySelector to format several metrics which match a resource selector
func (p *testingProvider) GetMetricBySelector(_ context.Context, namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	return p.metricsFor(namespace, selector, info, metricSelector)
}
