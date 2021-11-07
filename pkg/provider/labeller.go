package provider

import "context"

type Labeller interface {
	Labels(ctx context.Context, namespace string) (map[string]string, error)
}

/*
 * FakeLabeller can be used to fake labels applied on namespace to mark them valid/invalid for openfaas functions
 */
type FakeLabeller struct {
	FakeLabels map[string]string
}

func (s *FakeLabeller) Labels(ctx context.Context, namespace string) (map[string]string, error) {
	return s.FakeLabels, nil
}
