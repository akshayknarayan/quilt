package monitor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/minion/docker"
	"github.com/NetSys/quilt/stitch"

	log "github.com/Sirupsen/logrus"
)

type connection struct {
	from db.Container
	to   db.Container
}

// Monitor checks via ping that each connection outlined in the stitch is live.
func Monitor(conn db.Conn, dckr docker.Client) {
	for range time.Tick(15 * time.Second) {
		doMonitor(conn, dckr)
	}
}

func doMonitor(conn db.Conn, dckr docker.Client) {
	allContainers := conn.SelectFromContainer(nil)
	edges, containers, err := getState(conn)
	if err != nil {
		log.WithError(err).Warn("[monitor] Invalid spec.")
		return
	}

	var conns []connection
	for _, e := range edges {
		fr, err := nodeToContainer(e.From, containers)
		if err != nil {
			continue // container not on this minion
		}
		to, err := nodeToContainer(e.To, allContainers)
		if err != nil {
			log.WithError(err).Warn(fmt.Sprintf(
				"[monitor] Destintation container %s not found", e.To))
			continue
		}

		conns = append(conns, connection{
			from: fr,
			to:   to,
		})
	}

	for _, c := range conns {
		log.Info(fmt.Sprintf("[monitor] %d (%s) -> %d (%s) monitoring",
			c.from.StitchID, c.from.IP, c.to.StitchID, c.to.IP))
	}

	down := checkConnections(dckr, conns)
	for _, fail := range down {
		log.Warn(fmt.Sprintf("[monitor] %s -> %s unreachable!",
			fail.from.IP, fail.to.IP))
	}
}

func getState(conn db.Conn) ([]stitch.Edge, []db.Container, error) {
	var minion db.Minion
	for {
		var err error
		minion, err = conn.MinionSelf()
		if err != nil {
			log.WithError(err).Warn("[monitor] Could not find minion.")
		} else {
			break
		}

		time.Sleep(30 * time.Second)
	}

	dbcs := conn.SelectFromContainer(func(dbc db.Container) bool {
		return dbc.Minion == minion.PrivateIP
	})

	spec, err := stitch.New(minion.Spec)
	if err != nil {
		return []stitch.Edge{}, dbcs, err
	}

	graph, err := stitch.InitializeGraph(spec)
	if err != nil {
		return []stitch.Edge{}, dbcs, err
	}
	return graph.GetConnections(), dbcs, nil
}

func checkConnections(dk docker.Client, conns []connection) []connection {
	var down []connection
	for _, c := range conns {
		if ok := reach(dk, c.from, c.to); !ok {
			down = append(down, c)
		}
	}
	return down
}

func nodeToContainer(id string, containers []db.Container) (db.Container, error) {
	cid, err := strconv.Atoi(id)
	if err != nil {
		log.WithError(err).Warn("id should always be from a container id (int)")
		return db.Container{}, err
	}

	for _, cnt := range containers {
		if cnt.StitchID == cid {
			return cnt, nil
		}
	}

	return db.Container{}, fmt.Errorf("container not found: %s", id)
}

func reach(dk docker.Client, ctr db.Container, toReach db.Container) bool {
	return pingInNamespace(dk, ctr, toReach.IP)
}

func pingInNamespace(dk docker.Client, ctr db.Container, ip string) bool {
	dCont, err := dk.Get(ctr.DockerID)
	if err != nil {
		log.WithError(err).Warn("container not found")
		return false
	}
	out, exitCode, err := dk.ExecVerbose(dCont.Name[1:], "ping", "-q", "-c3", ip)
	if err != nil {
		log.WithError(err).Warn(string(out))
		return false
	}

	if exitCode == 0 {
		return true
	}
	log.Warn(out)
	return false
}
