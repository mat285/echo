package selector

// FromLabels returns a selector matching the labels
func FromLabels(labels map[string]string) string {
	preds := []string{}
	for k, v := range labels {
		if len(k) > 0 {
			preds = append(preds, Equals(k, v))
		}
	}
	return And(preds...)
}
