package raft

func (rf *Raft) to(state int) {
	rf.mu.Lock()
	rf.state = state
	rf.mu.Unlock()
}

func (rf *Raft) clearChan() {
	select {
	case <-rf.heartbeatChan:
	default:
	}
	select {
	case <-rf.voteResultChan:
	default:
	}
}


