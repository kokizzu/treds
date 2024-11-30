package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/pool"
	wal "github.com/hashicorp/raft-wal"
	"treds/commands"
	"treds/store"

	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	"github.com/panjf2000/gnet/v2"
)

const Snapshot = "SNAPSHOT"
const Restore = "RESTORE"

type BootStrapServer struct {
	ID   string
	Host string
	Port int
}

type Server struct {
	Addr string
	Port int

	tredsCommandRegistry commands.CommandRegistry

	*gnet.BuiltinEventEngine
	fsm               *TredsFsm
	raft              *raft.Raft
	id                string
	tcpConnectionPool pool.Pool
	raftApplyTimeout  time.Duration
}

func New(port, segmentSize int, bindAddr, advertiseAddr, serverId string, applyTimeout time.Duration, servers []BootStrapServer) (*Server, error) {

	commandRegistry := commands.NewRegistry()
	commands.RegisterCommands(commandRegistry)
	tredsStore := store.NewTredsStore()

	//TODO: Default config is good enough for now, but probably need to be tweaked
	config := raft.DefaultConfig()

	serverIdFileName := "server-id"

	if serverId == "" {
		// try reading from file
		if _, err := os.Stat(serverIdFileName); err == nil {
			// File exists, read the UUID
			fmt.Println("File found. Reading UUID from file...")
			data, readErr := os.ReadFile(serverIdFileName)
			if readErr != nil {
				fmt.Println("Error reading UUID from file:", err)
			}
			// Parse the UUID
			id, parseErr := uuid.Parse(string(data))
			if parseErr != nil {
				fmt.Println("Error parsing UUID:", parseErr)
			}
			fmt.Println("UUID read from file:", id)
			config.LocalID = raft.ServerID(id.String())

		} else if os.IsNotExist(err) {
			// File does not exist, generate a new UUID
			fmt.Println("File not found. Generating a new UUID...")
			id := uuid.New()

			// Write the UUID to the file
			err = os.WriteFile(serverIdFileName, []byte(id.String()), 0644)
			if err != nil {
				fmt.Println("Error writing UUID to file:", err)
			}
			fmt.Println("New UUID generated and written to file:", id)
			config.LocalID = raft.ServerID(id.String())
		} else {
			// Other errors (e.g., permission issues)
			fmt.Println("Error checking file:", err)
			id := serverId
			config.LocalID = raft.ServerID(id)
		}
	} else {
		// try reading from file
		if _, err := os.Stat(serverIdFileName); err == nil {
			// File exists, read the UUID
			fmt.Println("File found. Reading UUID from file...")
			data, readErr := os.ReadFile(serverIdFileName)
			if readErr != nil {
				fmt.Println("Error reading UUID from file:", err)
			}
			// Parse the UUID
			id, parseErr := uuid.Parse(string(data))
			if parseErr != nil {
				fmt.Println("Error parsing UUID:", parseErr)
			}
			if id.String() != serverId {
				return nil, fmt.Errorf("UUID does not match")
			}
			fmt.Println("UUID read from file:", id)
			config.LocalID = raft.ServerID(id.String())

		} else if os.IsNotExist(err) {
			// File does not exist, generate a new UUID
			fmt.Println("File not found. Generating a new UUID...")
			id := serverId

			// Write the UUID to the file
			err = os.WriteFile(serverIdFileName, []byte(id), 0644)
			if err != nil {
				fmt.Println("Error writing UUID to file:", err)
			}
			fmt.Println("New UUID generated and written to file:", id)
			config.LocalID = raft.ServerID(id)
		} else {
			// Other errors (e.g., permission issues)
			fmt.Println("Error checking file:", err)
			id := serverId
			config.LocalID = raft.ServerID(id)
		}
	}

	//This is the port used by raft for replication and such
	// We can keep it as a separate port or do multiplexing over TCP
	addr := fmt.Sprintf("%s:%d", bindAddr, 8300)

	transport, err := raft.NewTCPTransport(addr, &net.TCPAddr{IP: net.IP(advertiseAddr), Port: port}, 10, time.Second, os.Stdout)

	//TODO: do not panic
	if err != nil {
		return nil, err
	}

	// Use raft wal as a backend store for raft
	dir := filepath.Join("data", string(config.LocalID))

	err = os.MkdirAll(dir, fs.ModeDir|fs.ModePerm)
	if err != nil {

		return nil, err
	}

	w, err := wal.Open(dir, wal.WithSegmentSize(segmentSize))
	if err != nil {

		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore("data", 3, nil)
	if err != nil {
		return nil, err
	}

	fsm := NewTredsFsm(commandRegistry, tredsStore)
	r, err := raft.NewRaft(config, fsm, w, w, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	bootStrapServers := []raft.Server{{ID: config.LocalID, Address: raft.ServerAddress(addr), Suffrage: raft.Voter}}

	for _, server := range servers {
		bootStrapServers = append(bootStrapServers, raft.Server{
			ID:      raft.ServerID(server.ID),
			Address: raft.ServerAddress(fmt.Sprintf("%s:%d", server.Host, server.Port)),
		})
	}

	cluster := r.BootstrapCluster(raft.Configuration{Servers: bootStrapServers})

	err = cluster.Error()
	if err != nil {
		return nil, err
	}

	return &Server{
		Port:                 port,
		tredsCommandRegistry: commandRegistry,
		fsm:                  fsm,
		raft:                 r,
		id:                   string(config.LocalID),
		raftApplyTimeout:     applyTimeout,
	}, nil
}

func (ts *Server) OnBoot(_ gnet.Engine) gnet.Action {
	fmt.Println("Server started on", ts.Port)
	go func() {
		for {
			ts.fsm.tredsStore.CleanUpExpiredKeys()
			time.Sleep(100 * time.Millisecond)
		}
	}()
	go func() {
		for {
			leaderAddr, leaderId := ts.raft.LeaderWithID()
			if string(leaderId) == ts.id {
				return
			}

			// Create a factory() to be used with channel based pool
			factory := func() (net.Conn, error) { return net.Dial("tcp", string(leaderAddr)) }
			p, _ := pool.NewChannelPool(5, 30, factory)
			ts.tcpConnectionPool = p
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return gnet.None
}

func (ts *Server) OnTraffic(c gnet.Conn) gnet.Action {

	data, _ := c.Next(-1)
	inp := string(data)
	if inp == "" {
		err := fmt.Errorf("empty command")
		respErr := fmt.Sprintf("Error Executing command - %v\n", err.Error())
		_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(respErr), respErr)))
		if errConn != nil {
			fmt.Println("Error occurred writing to connection", errConn)
		}
		return gnet.None
	}

	if strings.ToUpper(inp) == Snapshot {

		// Only writes need to be forwarded to leader
		forwarded, rspFwd, err := ts.forwardRequest(data)
		if err != nil {
			respondErr(c, err)
			return gnet.None
		}

		// If request is forwarded we just send back the answer from the leader to the client
		// and stop processing
		if forwarded {
			_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(rspFwd), rspFwd)))
			if errConn != nil {
				fmt.Println("Error occurred writing to connection", errConn)
			}
			return gnet.None
		}

		future := ts.raft.Snapshot()
		if future.Error() != nil {
			respondErr(c, future.Error())
			return gnet.None
		}
		res := "OK"
		_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(res), res)))
		if errConn != nil {
			respondErr(c, errConn)
		}
		return gnet.None
	}

	if strings.ToUpper(strings.Split(inp, " ")[0]) == Restore {
		// Only writes need to be forwarded to leader
		forwarded, rspFwd, err := ts.forwardRequest(data)
		if err != nil {
			respondErr(c, err)
			return gnet.None
		}

		// If request is forwarded we just send back the answer from the leader to the client
		// and stop processing
		if forwarded {
			_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(rspFwd), rspFwd)))
			if errConn != nil {
				fmt.Println("Error occurred writing to connection", errConn)
			}
			return gnet.None
		}

		snapshotPath := strings.Split(inp, " ")[1]

		metaFile := filepath.Join(snapshotPath, "meta.json")

		// Read the file contents
		metaData, err := os.ReadFile(metaFile)
		if err != nil {
			fmt.Println("Error reading file:", err)
			respondErr(c, err)
			return gnet.None
		}

		// Unmarshal the JSON into the SnapshotMeta struct
		var metaSnapshot *raft.SnapshotMeta
		err = json.Unmarshal(metaData, &metaSnapshot)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			respondErr(c, err)
			return gnet.None
		}

		file, err := os.Open(filepath.Join(snapshotPath, "state.bin"))
		if err != nil {
			fmt.Println("Error opening file:", err)
			respondErr(c, err)
			return gnet.None
		}
		// Ensure the file is closed when done
		defer file.Close()

		// Since *os.File implements io.Reader, you can directly use it as an io.Reader
		var reader io.Reader = file

		err = ts.raft.Restore(metaSnapshot, reader, 2*time.Minute)
		if err != nil {
			respondErr(c, err)
			return gnet.None
		}
		res := "OK"
		_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(res), res)))
		if errConn != nil {
			respondErr(c, errConn)
		}
		return gnet.None
	}

	commandStringParts := parseCommand(inp)
	commandReg, err := ts.tredsCommandRegistry.Retrieve(strings.ToUpper(commandStringParts[0]))
	if err != nil {
		respondErr(c, err)
		return gnet.None
	}
	if commandReg.IsWrite {

		// Only writes need to be forwarded to leader
		forwarded, rspFwd, err := ts.forwardRequest(data)
		if err != nil {
			respondErr(c, err)
			return gnet.None
		}

		// If request is forwarded we just send back the answer from the leader to the client
		// and stop processing
		if forwarded {
			_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(rspFwd), rspFwd)))
			if errConn != nil {
				fmt.Println("Error occurred writing to connection", errConn)
			}
			return gnet.None
		}

		// Validation need to be done before raft Apply so an error is returned before persisting
		if err = commandReg.Validate(commandStringParts[1:]); err != nil {
			respondErr(c, err)
			return gnet.None
		}

		future := ts.raft.Apply(data, ts.raftApplyTimeout)

		if err := future.Error(); err != nil {
			respondErr(c, err)
			return gnet.None
		}
		rsp := future.Response()

		switch rsp.(type) {
		case error:
			err := rsp.(error)
			respondErr(c, err)
			return gnet.None
		default:
			res := "OK"
			_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(res), res)))
			if errConn != nil {
				fmt.Println("Error occurred writing to connection", errConn)
			}
		}
	} else {
		if err = commandReg.Validate(commandStringParts[1:]); err != nil {
			respondErr(c, err)
			return gnet.None
		}
		res := commandReg.Execute(commandStringParts[1:], ts.fsm.tredsStore)
		_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(res), res)))
		if errConn != nil {
			fmt.Println("Error occurred writing to connection", errConn)
		}
	}
	return gnet.None
}

func parseCommand(inp string) []string {
	commandString := strings.TrimSpace(inp)
	commandStringParts := strings.Split(commandString, " ")
	return commandStringParts
}

func respondErr(c gnet.Conn, err error) {
	respErr := fmt.Sprintf("Error Executing command - %v\n", err.Error())
	_, errConn := c.Write([]byte(fmt.Sprintf("%d\n%s", len(respErr), respErr)))
	if errConn != nil {
		fmt.Println("Error occurred writing to connection", errConn)
	}
}

func (ts *Server) OnClose(_ gnet.Conn, _ error) gnet.Action {
	return gnet.None
}

func (ts *Server) forwardRequest(data []byte) (bool, string, error) {
	// create a new channel based pool with an initial capacity of 5 and maximum
	// capacity of 30. The factory will create 5 initial connections and put it
	// into the pool.

	_, leaderId := ts.raft.LeaderWithID()

	if ts.id == string(leaderId) {
		return false, "", nil
	}

	conn, err := ts.tcpConnectionPool.Get()
	if err != nil {
		return false, "", nil
	}
	defer conn.Close()
	_, err = conn.Write(data)
	if err != nil {
		return false, "", nil
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, "", err
	}
	return true, line, nil
}
