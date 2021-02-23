package raft

import "time"

type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

type AppendEntriesReply struct {
	Term int
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.currentTerm
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.state = STATE_FOLLOWER
	}

	if args.Term == rf.currentTerm {
		rf.state = STATE_FOLLOWER
		rf.heartbeatChan <- *args
	}
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, replyChan chan AppendEntriesReply) bool {
	reply := AppendEntriesReply{}
	ok := rf.peers[server].Call("Raft.AppendEntries", args, &reply)
	replyChan <- reply
	return ok
}

func (rf *Raft) broadcastHeartbeat() {
	rf.mu.Lock()
	term := rf.currentTerm
	rf.mu.Unlock()
	reply := AppendEntriesReply{}
	replyChan := make(chan AppendEntriesReply, len(rf.peers)-1)
	for i := range rf.peers {
		if i != rf.me {
			go rf.sendAppendEntries(i, &AppendEntriesArgs{Term: term, LeaderId: rf.me}, replyChan)
		}
	}
	for i := 0; i < len(rf.peers)-1; i++ {
		select {
		case reply = <-replyChan:
			rf.mu.Lock()
			if reply.Term > rf.currentTerm {
				rf.currentTerm = reply.Term
				rf.votedFor = -1
				rf.state = STATE_FOLLOWER
				rf.mu.Unlock()
				break
			}
			rf.mu.Unlock()
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
}
