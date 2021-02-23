package raft

import (
	//	"bytes"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

const (
	STATE_FOLLOWER  int = 0
	STATE_LEADER    int = 1
	STATE_CANDITATE int = 2
)

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	state       int
	votedFor    int
	currentTerm int

	heartbeatChan   chan AppendEntriesArgs
	requestVoteChan chan RequestVoteArgs

	voteResultChan chan bool

	electionDuration time.Duration

	heartbeatTimeout  time.Duration
	heartbeatDuration time.Duration
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.state == STATE_LEADER
}

func (rf *Raft) persist() {

}

func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}

func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		rf.resetTimeout()
		rf.mu.Lock()
		state := rf.state
		rf.mu.Unlock()
		rf.clearChan()
		switch state {
		case STATE_FOLLOWER:
			select {
			case <-rf.heartbeatChan:
			case <-time.After(rf.heartbeatTimeout):
				rf.to(STATE_CANDITATE)
			}
		case STATE_CANDITATE:
			rf.mu.Lock()
			rf.currentTerm++
			rf.votedFor = rf.me
			rf.mu.Unlock()
			go rf.broadcastRequestVote()
			select {
			case <-rf.heartbeatChan:
			case voteResult := <-rf.voteResultChan:
				if voteResult {
					rf.to(STATE_LEADER)
					go rf.broadcastHeartbeat()
				}
			case <-time.After(rf.electionDuration):
			}
		case STATE_LEADER:
			select {
			case <-rf.heartbeatChan:
			case <-time.After(rf.heartbeatDuration):
				go rf.broadcastHeartbeat()
			}
		}
	}
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.state = STATE_FOLLOWER
	rf.heartbeatDuration = 100 * time.Millisecond
	rf.electionDuration = 1000 * time.Millisecond
	rf.readPersist(persister.ReadRaftState())
	rf.votedFor = -1
	rf.heartbeatChan = make(chan AppendEntriesArgs)
	rf.voteResultChan = make(chan bool)
	rf.currentTerm = 0
	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}
