package utils

import (
	"encoding/json"
	"fmt"
	"github.com/Superdanda/hade/framework/gin"
)

func ConvertToSpecificType[T any, I interface{}](items []I, convertFunc func(I) (T, bool)) ([]T, error) {
	specificItems := make([]T, 0, len(items))
	for i, item := range items {
		if specificItem, ok := convertFunc(item); ok {
			specificItems = append(specificItems, specificItem)
		} else {
			return nil, fmt.Errorf("invalid type at index %d", i)
		}
	}
	return specificItems, nil
}

func ConvertToAbstractNodes[T any, I interface{}](items []T, toInterface func(T) I) []I {
	nodes := make([]I, len(items))
	for i, item := range items {
		nodes[i] = toInterface(item)
	}
	return nodes
}

func QuickBind[T any](c *gin.Context) *T {
	var params T
	if err := c.ShouldBindJSON(&params); err != nil {
		return nil
	}
	return &params
}

func Convert(origin interface{}, target interface{}) error {
	data, err := json.Marshal(origin)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, target)
	if err != nil {
		return err
	}
	return nil
}
