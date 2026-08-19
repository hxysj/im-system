package utils

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

var idNode *snowflake.Node

func InitIdGenerator(nodeID int64) error {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("初始化发号器失败：%w",err)
	}
	idNode = node 
	return nil
}

func NextId()int64{
	if idNode == nil{
		panic("发号器未初始化")
	}

	return idNode.Generate().Int64()
}