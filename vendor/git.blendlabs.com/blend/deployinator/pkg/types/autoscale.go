package types

// AutoscaleMetricAggregation is the aggregation mode of the metric
type AutoscaleMetricAggregation string

// With AutoscaleMetricAggregationNone, kube compares the raw value of the metric
// to the threshold using `targetValue`. With AutoscaleMetricAggregationDivideByPod,
// kube divides the metric value by the number of pods using `targetAverageValue`.
const (
	AutoscaleMetricAggregationNone        AutoscaleMetricAggregation = "none"
	AutoscaleMetricAggregationDivideByPod AutoscaleMetricAggregation = "divideByPod"
)

// AutoscaleMetric defines external metric source for the autoscaler
type AutoscaleMetric struct {
	Name            string                     `json:"name" yaml:"name"`
	Selector        map[string]string          `json:"selector,omitempty" yaml:"selector,omitempty"`
	Threshold       string                     `json:"threshold" yaml:"threshold"`
	KubeAggregation AutoscaleMetricAggregation `json:"kubeAggregation" yaml:"kubeAggregation"`
}

// IsValid validates if the AutoscaleMetrics is valid
func (m AutoscaleMetric) IsValid() bool {
	return len(m.Name) > 0 && len(m.Threshold) > 0
}
