package compose_test

import (
	"log"
	"os"
	"testing"

	"github.com/bravetools/bravetools/platform"
	"github.com/bravetools/bravetools/shared"
)

func TestCompose(t *testing.T) {
	host := *platform.NewBraveHost()
	backend, err := platform.NewHostBackend(host)
	if err != nil {
		log.Fatal(err)
	}

	os.Chdir("python-multi-service")
	p := "brave-compose.yaml"

	composefile := shared.NewComposeFile()
	err = composefile.Load(p)
	if err != nil {
		log.Fatal("Failed to load compose file: ", err)
	}

	err = host.Compose(backend, composefile)
	if err != nil {
		log.Fatal(err)
	}

	//Cleanup
	for service := range composefile.Services {
		host.DeleteUnit(service)
		host.DeleteLocalImage(composefile.Services[service].Image)
	}

}
