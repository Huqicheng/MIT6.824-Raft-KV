package raft

import (
	"math/rand"
	"time"
)

type RequestVoteArgs struct {
	Term        int
	CandidateId int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.currentTerm

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.state = STATE_FOLLOWER
	}

	if args.Term == rf.currentTerm && (rf.votedFor == -1 || rf.votedFor == args.CandidateId) {
		reply.VoteGranted = true
		reply.Term = args.Term
		rf.votedFor = args.CandidateId
	} else {
		reply.VoteGranted = false
		reply.Term = args.Term
	}
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, replyChan chan RequestVoteReply) bool {
	reply := RequestVoteReply{}
	ok := rf.peers[server].Call("Raft.RequestVote", args, &reply)
	replyChan <- reply
	return ok
}

func (rf *Raft) broadcastRequestVote() {
	rf.mu.Lock()
	term := rf.currentTerm
	rf.mu.Unlock()
	reply := RequestVoteReply{}
	voteCnt := 1
	replyChan := make(chan RequestVoteReply, len(rf.peers)-1)
	for i := range rf.peers {
		if i != rf.me {
			go rf.sendRequestVote(i, &RequestVoteArgs{Term: term, CandidateId: rf.me}, replyChan)
		}
	}
	for i := 0; i <= len(rf.peers)-1; i++ {
		select {
		case reply = <-replyChan:
			rf.mu.Lock()
			if reply.Term > rf.currentTerm {
				rf.currentTerm = reply.Term
				rf.votedFor = -1
				rf.state = STATE_FOLLOWER
				rf.mu.Unlock()
				rf.voteResultChan <- false
				return
			}
			if reply.Term == rf.currentTerm && reply.VoteGranted {
				voteCnt++
			}
			if voteCnt > len(rf.peers)/2 {
				rf.mu.Unlock()
				rf.voteResultChan <- true
				return
			}
			rf.mu.Unlock()
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
	rf.voteResultChan <- false
}

func (rf *Raft) resetTimeout() {
	rf.heartbeatTimeout = time.Duration(rand.Intn(1000) + 100) * time.Millisecond
}
