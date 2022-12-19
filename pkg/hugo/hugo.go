package hugo

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

const tempDir = "/tmp/site/"
const configLocation = tempDir + "config.toml"

// updateHost takes writes the actual host (bucket url)
// to config.toml. It's used for links within the site.
func updateHost(host string) error {
	log.WithFields(log.Fields{
		"config home": configLocation,
	}).Debug("Updating Config")
	input, err := ioutil.ReadFile(configLocation)
	if err != nil {
		log.Error("Failed to find config.toml")
		return err
	}

	stringInput := string(input)

	stringInput = strings.Replace(stringInput, "http://localhost:1313/", host, 1)

	err = ioutil.WriteFile(configLocation, []byte(stringInput), 0775)
	if err != nil {
		log.Error("Failed to write config.toml back")
		return err
	}
	return nil
}

func list(path string) error {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": path,
		}).Error("Failed to list directory")
		return err
	}

	for _, f := range files {
		log.Debug(f.Name())
	}
	return nil
}

func copyHugo() error {
	source, err := os.Open("./hugo")
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create("/tmp/hugo")
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	err = os.Chmod("/tmp/hugo", 0775)
	if err != nil {
		return err
	}
	return nil
}

// Compile - Builds the Hugo static site
func Compile(host string) error {
	err := copyHugo()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to copy Hugo")
	}

	err = updateHost(host)
	if err != nil {
		return err
	}
	list("/tmp/")

	args := []string{"-v", "--source", tempDir, "--destination", tempDir + "public"}
	cmd := exec.Command("/tmp/hugo")
	cmd.Args = args

	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Hugo failed to run")
		return err
	}
	list("/tmp/site/public")
	log.Info("Successfully copiled app")
	return nil
}
