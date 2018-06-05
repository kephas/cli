package deployments

import (
	"encoding/json"
	"fmt"
	stdio "io"
	"strings"
	"time"

	"github.com/Scalingo/cli/config"
	"github.com/Scalingo/cli/debug"
	"github.com/Scalingo/go-scalingo"
	"golang.org/x/net/websocket"
	"gopkg.in/errgo.v1"
)

type StreamOpts struct {
	AppName      string
	DeploymentID string
}

type deployEvent struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type logData struct {
	Content string `json:"content"`
}

type statusData struct {
	Content string `json:"Status"`
}

// We can stream the deployment logs of an application, or we can stream the logs of a specific
// deployments.
// The StreamOpts.DeploymentID argument is optional.
func Stream(opts *StreamOpts) error {
	c := config.ScalingoClient()
	app, err := c.AppsShow(opts.AppName)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	debug.Println("Opening socket to: " + app.Links.DeploymentsStream)

	conn, err := c.DeploymentStream(app.Links.DeploymentsStream)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	var event deployEvent
	currentDeployment := &scalingo.Deployment{
		ID: opts.DeploymentID,
	}
	anyDeployment := currentDeployment.ID == ""
	statuses := map[string]string{}

	for {
		err := websocket.JSON.Receive(conn, &event)
		if err != nil {
			conn.Close()
			if err == stdio.EOF {
				debug.Println("Remote server broke the connection, reconnecting")
				for err != nil {
					conn, err = c.DeploymentStream(app.Links.DeploymentsStream)
					time.Sleep(time.Second * 1)
				}
				continue
			} else {
				return errgo.Mask(err, errgo.Any)
			}
		} else {
			switch event.Type {
			case "ping":
			case "log":
				// If we stream logs of a specific deployment and this event is not about this one
				if !anyDeployment && event.ID != currentDeployment.ID {
					continue
				}
				var logData logData
				err := json.Unmarshal(event.Data, &logData)
				if err != nil {
					config.C.Logger.Println(err)
					continue
				}
				fmt.Println("[LOG] " + strings.TrimSpace(logData.Content))
			case "status":
				// If we stream logs of a specific deployment and this event is not about this one
				if !anyDeployment && event.ID != currentDeployment.ID {
					continue
				}
				var statusData statusData
				err := json.Unmarshal(event.Data, &statusData)
				if err != nil {
					config.C.Logger.Println(err)
					continue
				}
				if statuses[event.ID] == "" {
					fmt.Println("[STATUS] New status: " + statusData.Content)
				} else {
					fmt.Println("[STATUS] New status: " + statuses[event.ID] + " →  " + statusData.Content)
				}
				statuses[event.ID] = statusData.Content

				if !anyDeployment && scalingo.IsFinishedString(scalingo.DeploymentStatus(statusData.Content)) {
					return nil
				}
			case "new":
				var newData map[string]*scalingo.Deployment
				err := json.Unmarshal(event.Data, &newData)
				if err != nil {
					config.C.Logger.Println(err)
					continue
				}
				newDeployment := newData["deployment"]

				if newDeployment.ID == currentDeployment.ID {
					currentDeployment = newDeployment
				}

				if anyDeployment || newDeployment.ID == currentDeployment.ID {
					fmt.Println("[NEW] New deploy: " + newDeployment.ID + " from " + newDeployment.User.Username)
				}
			}
		}
	}
}
