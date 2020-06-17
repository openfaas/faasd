package pkg

import (
	"log"
	"testing"
)

func Test_buildDeploymentOrder_ARequiresB(t *testing.T) {
	svcs := []Service{
		{
			Name:      "A",
			DependsOn: []string{"B"},
		},
		{
			Name:      "B",
			DependsOn: []string{},
		},
	}

	order := buildDeploymentOrder(svcs)

	if len(order) < len(svcs) {
		t.Fatalf("length of order too short: %d", len(order))
	}

	got := order[0]
	want := "B"
	if got != want {
		t.Fatalf("%s should be last to be installed, but was: %s", want, got)
	}
}

func Test_buildDeploymentOrder_ARequiresBAndC(t *testing.T) {
	svcs := []Service{
		{
			Name:      "A",
			DependsOn: []string{"B", "C"},
		},
		{
			Name:      "B",
			DependsOn: []string{},
		},
		{
			Name:      "C",
			DependsOn: []string{},
		},
	}

	order := buildDeploymentOrder(svcs)

	if len(order) < len(svcs) {
		t.Fatalf("length of order too short: %d", len(order))
	}

	a := indexStr(order, "a")
	b := indexStr(order, "b")
	c := indexStr(order, "c")

	if a > b {
		t.Fatalf("a should be after dependencies")
	}
	if a > c {
		t.Fatalf("a should be after dependencies")
	}

}

func Test_buildDeploymentOrder_ARequiresBRequiresC(t *testing.T) {
	svcs := []Service{
		{
			Name:      "A",
			DependsOn: []string{"B"},
		},
		{
			Name:      "B",
			DependsOn: []string{"C"},
		},
		{
			Name:      "C",
			DependsOn: []string{},
		},
	}

	order := buildDeploymentOrder(svcs)

	if len(order) < len(svcs) {
		t.Fatalf("length of order too short: %d", len(order))
	}

	got := order[0]
	want := "C"
	if got != want {
		t.Fatalf("%s should be last to be installed, but was: %s", want, got)
	}
	got = order[1]
	want = "B"
	if got != want {
		t.Fatalf("%s should be last to be installed, but was: %s", want, got)
	}
	got = order[2]
	want = "A"
	if got != want {
		t.Fatalf("%s should be last to be installed, but was: %s", want, got)
	}
}

func Test_buildDeploymentOrderCircularARequiresBRequiresA(t *testing.T) {
	svcs := []Service{
		{
			Name:      "A",
			DependsOn: []string{"B"},
		},
		{
			Name:      "B",
			DependsOn: []string{"A"},
		},
	}

	defer func() { recover() }()

	buildDeploymentOrder(svcs)

	t.Fatalf("did not panic as expected")
}

func Test_buildDeploymentOrderComposeFile(t *testing.T) {
	// svcs := []Service{}
	file, err := LoadComposeFileWithArch("../", "docker-compose.yaml", func() (string, string) {
		return "x86_64", "Linux"
	})

	if err != nil {
		t.Fatalf("unable to load compose file: %s", err)
	}

	svcs, err := ParseCompose(file)
	if err != nil {
		t.Fatalf("unable to parse compose file: %s", err)
	}

	for _, s := range svcs {
		log.Printf("Service: %s\n", s.Name)
		for _, d := range s.DependsOn {
			log.Printf("Link: %s => %s\n", s.Name, d)
		}
	}

	order := buildDeploymentOrder(svcs)

	if len(order) < len(svcs) {
		t.Fatalf("length of order too short: %d", len(order))
	}

	queueWorker := indexStr(order, "queue-worker")
	nats := indexStr(order, "nats")
	gateway := indexStr(order, "gateway")
	prometheus := indexStr(order, "prometheus")

	if prometheus > gateway {
		t.Fatalf("Prometheus order was after gateway, and should be before")
	}
	if nats > gateway {
		t.Fatalf("NATS order was after gateway, and should be before")
	}
	if nats > queueWorker {
		t.Fatalf("NATS order was after queue-worker, and should be before")
	}
}

func Test_buildDeploymentOrderOpenFaaS(t *testing.T) {
	svcs := []Service{
		{
			Name:      "queue-worker",
			DependsOn: []string{"nats"},
		},
		{
			Name:      "prometheus",
			DependsOn: []string{},
		},
		{
			Name:      "gateway",
			DependsOn: []string{"prometheus", "nats", "basic-auth-plugin"},
		},
		{
			Name:      "basic-auth-plugin",
			DependsOn: []string{},
		},
		{
			Name:      "nats",
			DependsOn: []string{},
		},
	}

	order := buildDeploymentOrder(svcs)

	if len(order) < len(svcs) {
		t.Fatalf("length of order too short: %d", len(order))
	}

	queueWorker := indexStr(order, "queue-worker")
	nats := indexStr(order, "nats")
	gateway := indexStr(order, "gateway")
	prometheus := indexStr(order, "prometheus")

	if prometheus > gateway {
		t.Fatalf("Prometheus order was after gateway, and should be before")
	}
	if nats > gateway {
		t.Fatalf("NATS order was after gateway, and should be before")
	}
	if nats > queueWorker {
		t.Fatalf("NATS order was after queue-worker, and should be before")
	}
}

func indexStr(st []string, t string) int {
	for n, s := range st {
		if s == t {
			return n
		}

	}
	return -1
}
