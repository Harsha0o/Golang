package main

import (
	//"encoding/json"
	"fmt"
	"net/http"

	//"os"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

type Asset struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func main() {
	r := gin.Default()
	r.POST("/createAsset", createAsset)
	r.POST("/updateAsset", updateAsset)
	r.GET("/queryAsset/:id", queryAsset)
	r.GET("/getAssetHistory/:id", getAssetHistory)
	r.Run(":8080")
}

func getGateway() (*client.Gateway, error) {
	wallet, err := identity.NewFileSystemWallet("wallet") // Check this method
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	gateway, err := client.Connect(client.WithIdentity(wallet, "user1"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	return gateway, nil
}

func createAsset(c *gin.Context) {
	var asset Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gateway, err := getGateway()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer gateway.Close()

	network, err := gateway.GetNetwork("mychannel") // Fixed assignment
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	contract := network.GetContract("assetContract")

	_, err = contract.SubmitTransaction("CreateAsset", asset.ID, asset.Name, fmt.Sprintf("%d", asset.Value))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset created successfully"})
}

// Implement other functions similarly (updateAsset, queryAsset, getAssetHistory)
