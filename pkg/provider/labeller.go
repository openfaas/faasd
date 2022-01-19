package provider

import "context"

// Labeller can return labels for a namespace from containerd.
type Labeller interface {
	Labels(ctx context.Context, namespace string) (map[string]string, error)
}

//
// FakeLabeller can be used to fake labels applied on namespace to mark
// them valid/invalid for openfaas functions
type FakeLabeller struct {
	labels map[string]string
}

func NewFakeLabeller(labels map[string]string) Labeller {
	return &FakeLabeller{
		labels: labels,
	}
}

func (s *FakeLabeller) Labels(ctx context.Context, namespace string) (map[string]string, error) {
	return s.labels, nil
}
