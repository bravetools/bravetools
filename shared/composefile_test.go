package shared

import "testing"

//
//                  ┌───────┐
//                ┌─┤  api  │
// ┌───────┐      │ └───────┘
// │  db   │◄─────┤
// └───────┘      │ ┌───────┐
//                └─┤ auth  │
//                  └───────┘
func TestTopologicalOrdering_first(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {Depends: []string{"db"}},
			"auth": {Depends: []string{"db"}},
			"db":   {},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()

	expectedFirstEle := "db"
	if err != nil {
		t.Fatalf("expected to be able to order graph")
	}
	actualFirstEle := ordering[0]

	if expectedFirstEle != actualFirstEle {
		t.Errorf("expected first element to be %q, found %q", expectedFirstEle, actualFirstEle)
	}
}

//
// ┌───────┐
// │ auth  │◄─┐
// └───────┘  │     ┌───────┐
//            ├─────┤  api  │
// ┌───────┐  │     └───────┘
// │  db   │◄─┘
// └───────┘
func TestTopologicalOrdering_end(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {Depends: []string{"db", "auth"}},
			"auth": {},
			"db":   {},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()

	expectedLastEle := "api"
	if err != nil {
		t.Fatalf("expected to be able to order graph")
	}
	actualLastEle := ordering[len(ordering)-1]

	if expectedLastEle != actualLastEle {
		t.Errorf("expected first element to be %q, found %q", expectedLastEle, actualLastEle)
	}
}

//
// ┌──────┐    ┌───────┐    ┌──────┐
// │ db   │◄───┤  api  │◄───┤ auth │
// └──────┘    └───────┘    └──────┘
func TestTopologicalOrdering_exactOrder(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {Depends: []string{"db"}},
			"auth": {Depends: []string{"api"}},
			"db":   {},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()
	if err != nil {
		t.Fatalf("expected to be able to order graph")
	}

	expectedOrdering := []string{"db", "api", "auth"}

	for i := range expectedOrdering {
		expectedEle := expectedOrdering[i]
		actualEle := ordering[i]

		if expectedEle != actualEle {
			t.Errorf("expected ordering %q, actual ordering %q", expectedOrdering, ordering)
		}
	}
}

//
// ┌──────┐    ┌───────┐    ┌──────┐
// │ auth │◄───┤  api  │◄───┤ db   │
// └──────┘    └───────┘    └──────┘
func TestTopologicalOrdering_exactOrder2(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {Depends: []string{"auth"}},
			"auth": {},
			"db":   {Depends: []string{"api"}},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()
	if err != nil {
		t.Fatalf("expected to be able to order graph")
	}

	expectedOrdering := []string{"auth", "api", "db"}

	for i := range expectedOrdering {
		expectedEle := expectedOrdering[i]
		actualEle := ordering[i]

		if expectedEle != actualEle {
			t.Errorf("expected ordering %q, actual ordering %q", expectedOrdering, ordering)
		}
	}
}

//
// ┌──────┐    ┌───────┐    ┌──────┐
// │ api  │◄───┤ auth  │◄───┤  db  │
// └──────┘    └───────┘    └──────┘
func TestTopologicalOrdering_exactOrder3(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {},
			"auth": {Depends: []string{"api"}},
			"db":   {Depends: []string{"auth"}},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()
	if err != nil {
		t.Fatalf("expected to be able to order graph")
	}

	expectedOrdering := []string{"api", "auth", "db"}

	for i := range expectedOrdering {
		expectedEle := expectedOrdering[i]
		actualEle := ordering[i]

		if expectedEle != actualEle {
			t.Errorf("expected ordering %q, actual ordering %q", expectedOrdering, ordering)
		}
	}
}

//
// ┌───────┐
// │ auth  │◄─┐
// └───┬───┘  │     ┌───────┐
//     │      ├─────┤  api  │
// ┌───▼───┐  │     └───────┘
// │  db   │◄─┘
// └───────┘
func TestTopologicalOrdering_exactOrder4(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]*ComposeService{
			"api":  {Depends: []string{"db", "auth"}},
			"auth": {Depends: []string{"db"}},
			"db":   {},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()
	if err != nil {
		t.Fatalf("expected to be able to order graph")
	}
	expectedOrdering := []string{"db", "auth", "api"}

	for i := range expectedOrdering {
		expectedEle := expectedOrdering[i]
		actualEle := ordering[i]

		if expectedEle != actualEle {
			t.Errorf("expected ordering %q, actual ordering %q", expectedOrdering, ordering)
		}
	}
}

//
// ┌───────┐
// │ auth  │
// └──▲──┬─┘
//    │  │
//    │  │
// ┌──┴──▼─┐    ┌──────┐
// │  api  │    │  db  │
// └───────┘    └──────┘
func TestTopologicalOrdering_directCycle(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {Depends: []string{"auth"}},
			"auth": {Depends: []string{"api"}},
			"db":   {},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()
	if err == nil {
		t.Errorf("expected to fail sorting due to cycle, found ordering: %q", ordering)
	}
}

//
//      ┌────────┐
//      │  auth  │
//      └─▲───┬──┘
//        │   │
//     ┌──┘   └────┐
//     │           │
// ┌───┴───┐    ┌──▼───┐
// │  api  │◄───┤  db  │
// └───────┘    └──────┘
func TestTopologicalOrdering_indirectCycle(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {Depends: []string{"auth"}},
			"auth": {Depends: []string{"db"}},
			"db":   {Depends: []string{"api"}},
		},
	}

	ordering, err := composeFile.TopologicalOrdering()
	if err == nil {
		t.Errorf("expected to fail sorting due to cycle, found ordering: %q", ordering)
	}
}

//
//
// ┌──────┐    ┌───────┐    ┌──────┐
// │ auth │    |  api  │    |  db  │
// └──────┘    └───────┘    └──────┘
func TestTopologicalOrdering_noDeps(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{
			"api":  {},
			"auth": {},
			"db":   {},
		},
	}

	_, err := composeFile.TopologicalOrdering()
	if err != nil {
		t.Errorf("expected no errors in graph with no deps")
	}
}

//
func TestTopologicalOrdering_noServices(t *testing.T) {
	composeFile := ComposeFile{
		Services: map[string]ComposeService{},
	}

	_, err := composeFile.TopologicalOrdering()
	if err != nil {
		t.Errorf("expected no errors in composeFile with no services")
	}
}
