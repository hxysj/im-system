package utils

import (
	"fmt"
	"sync"
	"time"
)

var (
	idNode     int64
	idSequence int64
	lastSecond int64
	idReady    bool
	idMutex    sync.Mutex
	idEpoch    int64 = 1735689600 // 2025-01-01 00:00:00 UTC
)

func InitIdGenerator(nodeID int64) error {
	if nodeID < 0 || nodeID > 899 {
		return fmt.Errorf("发号节点必须在 0 到 899 之间")
	}
	idMutex.Lock()
	idNode = nodeID
	lastSecond = time.Now().Unix() - idEpoch
	idSequence = time.Now().UnixNano() % 10000
	idReady = true
	idMutex.Unlock()
	return nil
}

func NextId() int64 {
	idMutex.Lock()
	defer idMutex.Unlock()

	if !idReady {
		panic("发号器未初始化")
	}

	second := time.Now().Unix() - idEpoch
	if second < 0 {
		panic("系统时间早于发号器起始时间")
	}
	if second == lastSecond {
		idSequence++
		if idSequence >= 10000 {
			for second <= lastSecond {
				time.Sleep(time.Millisecond)
				second = time.Now().Unix() - idEpoch
			}
			idSequence = 0
		}
	} else {
		idSequence = 0
	}
	lastSecond = second

	// 秒级时间、节点和序列组合后保持在 JavaScript 安全整数范围内。
	return second*1_000_000 + idNode*10_000 + idSequence
}
