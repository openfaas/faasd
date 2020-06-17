package handlers

import (
	"fmt"
	"testing"
)

func Test_BuildLabelsAndAnnotationsFromServiceSpec_Annotations(t *testing.T) {
	container := map[string]string{
		"qwer": "ty",
		"dvor": "ak",
		fmt.Sprintf("%scurrent-time", annotationLabelPrefix): "5 Nov 20:10:20 PST 1955",
		fmt.Sprintf("%sfuture-time", annotationLabelPrefix):  "21 Oct 20:10:20 PST 2015",
	}

	labels, annotation := buildLabelsAndAnnotations(container)

	if len(labels) != 2 {
		t.Errorf("want: %d labels got: %d", 2, len(labels))
	}

	if len(annotation) != 2 {
		t.Errorf("want: %d annotation got: %d", 1, len(annotation))
	}

	if _, ok := annotation["current-time"]; !ok {
		t.Errorf("want: '%s' entry in annotation map got: key not found", "current-time")
	}
}
