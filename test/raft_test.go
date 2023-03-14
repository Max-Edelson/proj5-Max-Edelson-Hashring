package SurfTest

import (
	"cse224/proj5/pkg/surfstore"
	"fmt"
	"os"
	"testing"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func TestRaftLogsCorrectlyOverwritten(t *testing.T) {
	t.Logf("leader1 gets several requests while all other nodes are crashed. leader1 crashes. all other nodes are restored. leader2 gets a request. leader1 is restored.")
	cfgPath := "./config_files/3nodes.txt"
	test := InitTest(cfgPath)
	defer EndTest(test)
	fmt.Printf("Set leader to server 0\n")
	test.Clients[0].SetLeader(test.Context, &emptypb.Empty{})
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	worker1 := InitDirectoryWorker("test0", SRC_PATH)
	worker2 := InitDirectoryWorker("test1", SRC_PATH)
	defer worker1.CleanUp()
	defer worker2.CleanUp()

	//clients add different files
	file1 := "multi_file1.txt"
	file2 := "multi_file1.txt"
	err := worker1.AddFile(file1)
	if err != nil {
		t.FailNow()
	}
	err = worker2.AddFile(file2)
	if err != nil {
		t.FailNow()
	}
	err = worker2.UpdateFile(file2, "update text")
	if err != nil {
		t.FailNow()
	}

	//client1 syncs
	//fmt.Printf("Start Sync\n")
	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	//fmt.Printf("End Sync\n")

	//fmt.Printf("Start heartbeat\n")
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	//fmt.Printf("End heartbeat\n")
	//fmt.Printf("Crash server 0\n")
	test.Clients[0].Crash(test.Context, &emptypb.Empty{})
	test.Clients[1].SetLeader(test.Context, &emptypb.Empty{})

	//fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	//fmt.Printf("End heartbeat\n")

	//client2 syncs
	//fmt.Printf("Start Sync\n")
	err = SyncClient("localhost:8080", "test1", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed: %s\n", err.Error())
	}
	//fmt.Printf("End Sync\n")

	//fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	//fmt.Printf("End heartbeat\n")

	//client1 syncs
	//fmt.Printf("Start Sync\n")
	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	//fmt.Printf("End Sync\n")

	//fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	//fmt.Printf("End heartbeat\n")
	//fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	//fmt.Printf("End heartbeat\n")

	workingDir, _ := os.Getwd()

	//check client1
	_, err = os.Stat(workingDir + "/test0/" + META_FILENAME)
	if err != nil {
		t.Fatalf("Could not find meta file for client1")
	}

	fileMeta1, err := LoadMetaFromDB(workingDir + "/test0/")
	if err != nil {
		t.Fatalf("Could not load meta file for client1")
	}
	if len(fileMeta1) != 1 {
		t.Fatalf("Wrong number of entries in client1 meta file")
	}
	if fileMeta1 == nil || fileMeta1[file1].Version != 1 {
		t.Fatalf("Wrong version for file1 in client1 metadata.")
	}

	c, e := SameFile(workingDir+"/test0/multi_file1.txt", SRC_PATH+"/multi_file1.txt")
	if e != nil {
		t.Fatalf("Could not read files in client base dirs.")
	}
	if !c {
		t.Fatalf("file1 should not change at client1")
	}

	//check client2
	_, err = os.Stat(workingDir + "/test1/" + META_FILENAME)
	if err != nil {
		t.Fatalf("Could not find meta file for client2")
	}

	fileMeta2, err := LoadMetaFromDB(workingDir + "/test1/")
	if err != nil {
		t.Fatalf("Could not load meta file for client2")
	}
	if len(fileMeta2) != 1 {
		t.Fatalf("Wrong number of entries in client2 meta file")
	}
	if fileMeta2 == nil || fileMeta2[file1].Version != 1 {
		t.Fatalf("Wrong version for file1 in client2 metadata.")
	}

	c, e = SameFile(workingDir+"/test1/multi_file1.txt", SRC_PATH+"/multi_file1.txt")
	if e != nil {
		t.Fatalf("Could not read files in client base dirs.")
	}
	if !c {
		t.Fatalf("wrong file2 contents at client2")
	}
}

func TestRaftUpdateTwice(t *testing.T) {
	t.Logf("leader1 gets a request. leader1 gets another request.")
	cfgPath := "./config_files/3nodes.txt"
	test := InitTest(cfgPath)
	defer EndTest(test)
	fmt.Printf("Set leader to server 0\n")
	test.Clients[0].SetLeader(test.Context, &emptypb.Empty{})
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	worker1 := InitDirectoryWorker("test0", SRC_PATH)
	defer worker1.CleanUp()

	//clients add different files
	file1 := "multi_file1.txt"
	err := worker1.AddFile(file1)
	if err != nil {
		t.FailNow()
	}

	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	fmt.Printf("Start heartbeat\n")
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	err = worker1.UpdateFile(file1, "update text")
	if err != nil {
		t.FailNow()
	}

	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	fmt.Printf("Start heartbeat\n")
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	workingDir, _ := os.Getwd()

	//check client1
	_, err = os.Stat(workingDir + "/test0/" + META_FILENAME)
	if err != nil {
		t.Fatalf("Could not find meta file for client1")
	}

	fileMeta1, err := LoadMetaFromDB(workingDir + "/test0/")
	if err != nil {
		t.Fatalf("Could not load meta file for client1")
	}
	fmt.Printf("fileMeta1: %v\n", fileMeta1[file1])
	if len(fileMeta1) != 1 {
		t.Fatalf("Wrong number of entries in client1 meta file")
	}
	if fileMeta1 == nil || fileMeta1[file1].Version != 2 {
		t.Fatalf("Wrong version for file1 in client1 metadata.")
	}

	c, e := SameFile(workingDir+"/test0/multi_file1.txt", workingDir+"/test0/multi_file1.txt")
	if e != nil {
		t.Fatalf("Could not read files in client base dirs.")
	}
	if !c {
		t.Fatalf("file1 should not change at client1")
	}
}

func TestRaftLogsConsistent(t *testing.T) {
	t.Logf("leader1 gets a request while a minority of the cluster is down. leader1 crashes. the other crashed nodes are restored. leader2 gets a request. leader1 is restored.")
	cfgPath := "./config_files/3nodes.txt"
	test := InitTest(cfgPath)
	defer EndTest(test)
	fmt.Printf("Set leader to server 0\n")
	test.Clients[0].SetLeader(test.Context, &emptypb.Empty{})
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	worker1 := InitDirectoryWorker("test0", SRC_PATH)
	worker2 := InitDirectoryWorker("test1", SRC_PATH)
	defer worker1.CleanUp()
	defer worker2.CleanUp()

	fmt.Printf("Crash server 1\n")
	test.Clients[1].Crash(test.Context, &emptypb.Empty{})

	//clients add different files
	file1 := "multi_file1.txt"
	file2 := "multi_file1.txt"
	err := worker1.AddFile(file1)
	if err != nil {
		t.FailNow()
	}

	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	fmt.Printf("Start heartbeat\n")
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	fmt.Printf("Crash server 0\n")
	test.Clients[0].Crash(test.Context, &emptypb.Empty{})

	fmt.Printf("Restore server 1\n")
	test.Clients[1].Restore(test.Context, &emptypb.Empty{})

	fmt.Printf("Set leader to server 1\n")
	test.Clients[1].SetLeader(test.Context, &emptypb.Empty{})
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	err = worker1.AddFile(file2)
	if err != nil {
		t.FailNow()
	}

	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	test.Clients[0].Restore(test.Context, &emptypb.Empty{})

	fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")
}

func TestRaftNewLeaderPushesUpdates(t *testing.T) {
	t.Logf("leader1 gets a request while the majority of the cluster is down. leader1 crashes. the other nodes come back. leader2 is elected\n")
	fmt.Printf("Send heartbeat from server 0")
	cfgPath := "./config_files/5nodes.txt"
	test := InitTest(cfgPath)
	defer EndTest(test)
	fmt.Printf("Set leader to server 0\n")
	test.Clients[0].SetLeader(test.Context, &emptypb.Empty{})
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	worker1 := InitDirectoryWorker("test0", SRC_PATH)
	defer worker1.CleanUp()

	fmt.Printf("Crash server 2,3,4\n")
	test.Clients[2].Crash(test.Context, &emptypb.Empty{})
	test.Clients[3].Crash(test.Context, &emptypb.Empty{})
	test.Clients[4].Crash(test.Context, &emptypb.Empty{})

	//clients add different files
	file1 := "multi_file1.txt"
	err := worker1.AddFile(file1)
	if err != nil {
		t.FailNow()
	}

	/*err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}*/

	fmt.Printf("Upload file to server 0.\n")
	hashlist := make([]string, 0)
	hashlist = append(hashlist, "9193460d4355655100cced0da2455725954377e10edfa83f18d347a10e2e2628")
	meta := surfstore.FileMetaData{Filename: "multi_file1.txt", Version: 1, BlockHashList: hashlist}
	test.Clients[0].UpdateFile(test.Context, &meta)

	fmt.Printf("Start heartbeat\n")
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	fmt.Printf("Crash server 0\n")
	test.Clients[0].Crash(test.Context, &emptypb.Empty{})

	fmt.Printf("Restore servers 2, 3, 4\n")
	test.Clients[2].Restore(test.Context, &emptypb.Empty{})
	test.Clients[3].Restore(test.Context, &emptypb.Empty{})
	test.Clients[4].Restore(test.Context, &emptypb.Empty{})

	fmt.Printf("Set leader to server 1\n")
	test.Clients[1].SetLeader(test.Context, &emptypb.Empty{})
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	for idx, server := range test.Clients {
		/*if idx == 0 {
			continue
		}*/
		state, _ := server.GetInternalState(test.Context, &emptypb.Empty{})
		if state == nil {
			t.Fatalf("Could not get state for server %d", idx)
		}
		/*if state.Term != int64(2) {
			t.Fatalf("Server should be in term %d", 2)
		}*/
		if len(state.Log) != 0 {
			t.Fatalf("Server %d has log length %d instead of 0", idx, len(state.Log))
		}
	}
}

func TestSyncTwoClientsFileUpdateLeaderFailure(t *testing.T) {
	t.Logf("client1 syncs with file1. client2 syncs. leader change. client2 syncs with file1 (different content). client1 syncs again.")
	cfgPath := "./config_files/3nodes.txt"
	test := InitTest(cfgPath)
	defer EndTest(test)
	fmt.Printf("Set leader to server 0\n")
	test.Clients[0].SetLeader(test.Context, &emptypb.Empty{})
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	worker1 := InitDirectoryWorker("test0", SRC_PATH)
	worker2 := InitDirectoryWorker("test1", SRC_PATH)
	defer worker1.CleanUp()
	defer worker2.CleanUp()

	//clients add different files
	file1 := "multi_file1.txt"
	file2 := "multi_file1.txt"
	err := worker1.AddFile(file1)
	if err != nil {
		t.FailNow()
	}
	err = worker2.AddFile(file2)
	if err != nil {
		t.FailNow()
	}

	//client1 syncs
	//fmt.Printf("Start Sync\n")
	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	//fmt.Printf("End Sync\n")

	//fmt.Printf("Start heartbeat\n")
	test.Clients[0].SendHeartbeat(test.Context, &emptypb.Empty{})
	//fmt.Printf("End heartbeat\n")

	//client2 syncs
	//fmt.Printf("Start Sync\n")
	err = SyncClient("localhost:8080", "test1", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed: %s\n", err.Error())
	}
	fmt.Printf("End Sync\n")

	fmt.Printf("Crash server 0\n")
	test.Clients[0].Crash(test.Context, &emptypb.Empty{})
	test.Clients[1].SetLeader(test.Context, &emptypb.Empty{})

	fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	err = worker2.UpdateFile(file2, "update text")
	if err != nil {
		t.FailNow()
	}

	fmt.Printf("=============== Sync updated file ===============\n\n")
	//client2 syncs
	fmt.Printf("Start Sync\n")
	err = SyncClient("localhost:8080", "test1", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed: %s\n", err.Error())
	}
	fmt.Printf("End Sync\n")

	fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	//client1 syncs
	fmt.Printf("Start Sync\n")
	err = SyncClient("localhost:8080", "test0", BLOCK_SIZE, cfgPath)
	if err != nil {
		t.Fatalf("Sync failed")
	}
	fmt.Printf("End Sync\n")

	fmt.Printf("Start heartbeat\n")
	test.Clients[1].SendHeartbeat(test.Context, &emptypb.Empty{})
	fmt.Printf("End heartbeat\n")

	workingDir, _ := os.Getwd()

	//check client1
	_, err = os.Stat(workingDir + "/test0/" + META_FILENAME)
	if err != nil {
		t.Fatalf("Could not find meta file for client1")
	}

	fileMeta1, err := LoadMetaFromDB(workingDir + "/test0/")
	if err != nil {
		t.Fatalf("Could not load meta file for client1")
	}
	fmt.Printf("fileMeta1: %v\n", fileMeta1[file1])
	if len(fileMeta1) != 1 {
		t.Fatalf("Wrong number of entries in client1 meta file")
	}
	if fileMeta1 == nil || fileMeta1[file1].Version != 2 {
		t.Fatalf("Wrong version for file1 in client1 metadata.")
	}

	c, e := SameFile(workingDir+"/test0/multi_file1.txt", workingDir+"/test1/multi_file1.txt")
	if e != nil {
		t.Fatalf("Could not read files in client base dirs.")
	}
	if !c {
		t.Fatalf("file1 should not change at client1")
	}

	//check client2
	_, err = os.Stat(workingDir + "/test1/" + META_FILENAME)
	if err != nil {
		t.Fatalf("Could not find meta file for client2")
	}

	fileMeta2, err := LoadMetaFromDB(workingDir + "/test1/")
	if err != nil {
		t.Fatalf("Could not load meta file for client2")
	}
	fmt.Printf("fileMeta2: %v\n", fileMeta2[file1])
	if len(fileMeta2) != 1 {
		t.Fatalf("Wrong number of entries in client2 meta file")
	}
	if fileMeta2 == nil || fileMeta2[file1].Version != 2 {
		t.Fatalf("Wrong version for file1 in client2 metadata.")
	}

	c, e = SameFile(workingDir+"/test1/multi_file1.txt", workingDir+"/test1/multi_file1.txt")
	if e != nil {
		t.Fatalf("Could not read files in client base dirs.")
	}
	if !c {
		t.Fatalf("wrong file2 contents at client2")
	}
}

func TestRaftSetLeader(t *testing.T) {
	//Setup
	cfgPath := "./config_files/3nodes.txt"
	test := InitTest(cfgPath)
	defer EndTest(test)

	// TEST
	leaderIdx := 0
	test.Clients[leaderIdx].SetLeader(test.Context, &emptypb.Empty{})

	// heartbeat
	for _, server := range test.Clients {
		server.SendHeartbeat(test.Context, &emptypb.Empty{})
	}

	//time.Sleep(3)
	for idx, server := range test.Clients {
		// all should have the leaders term
		state, _ := server.GetInternalState(test.Context, &emptypb.Empty{})
		if state == nil {
			t.Fatalf("Could not get state")
		}
		if state.Term != int64(1) {
			t.Fatalf("Server %d should be in term %d", idx, 1)
		}
		if idx == leaderIdx {
			// server should be the leader
			if !state.IsLeader {
				t.Fatalf("Server %d should be the leader", idx)
			}
		} else {
			// server should not be the leader
			if state.IsLeader {
				t.Fatalf("Server %d should not be the leader", idx)
			}
		}
	}

	leaderIdx = 2
	test.Clients[leaderIdx].SetLeader(test.Context, &emptypb.Empty{})

	// heartbeat
	for _, server := range test.Clients {
		server.SendHeartbeat(test.Context, &emptypb.Empty{})
	}

	for idx, server := range test.Clients {
		// all should have the leaders term
		state, _ := server.GetInternalState(test.Context, &emptypb.Empty{})
		if state == nil {
			t.Fatalf("Could not get state")
		}
		if state.Term != int64(2) {
			t.Fatalf("Server should be in term %d", 2)
		}
		if idx == leaderIdx {
			// server should be the leader
			if !state.IsLeader {
				t.Fatalf("Server %d should be the leader", idx)
			}
		} else {
			// server should not be the leader
			if state.IsLeader {
				t.Fatalf("Server %d should not be the leader", idx)
			}
		}
	}
}
