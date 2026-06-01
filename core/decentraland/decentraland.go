package decentraland

import (
	"context"
	"decentraland-data-downloader-v4/packages/database"
	"errors"
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	Decentraland = "decentraland"
)

const (
	EthereumBlockchain = "ethereum"
)

const (
	MANATokenSymbol = "MANA"
)

const (
	AssetTypeLand   = "parcel"
	AssetTypeEstate = "estate"
)

const (
	AssetAttributeNameSize  = "size"
	AssetAttributeNameOwner = "owner"
	AssetAttributeNameLands = "parcels"
)

const (
	AssetAttributeDataTypeInteger     = "integer"
	AssetAttributeDataTypeString      = "string"
	AssetAttributeDataTypeBoolean     = "boolean"
	AssetAttributeDataTypeStringArray = "string[]"
	AssetAttributeDataTypeAddress     = "address"
)

type DclInfoAssetAttribute struct {
	AttributeName  string         `bson:"attribute_name,omitempty"`
	DataType       string         `bson:"data_type,omitempty"`
	DataTypeParams map[string]any `bson:"data_type_params,omitempty"`
}

type DclInfoAsset struct {
	Blockchain string `bson:"blockchain,omitempty"`
	Key        string `bson:"key,omitempty"`
	Contract   string `bson:"symbol,omitempty"`
	Name       string `bson:"name,omitempty"`
	AssetType  string `bson:"asset_type,omitempty"`
	Attributes []DclInfoAssetAttribute
}

type DclInfoLogTopic struct {
	Blockchain  string   `bson:"blockchain,omitempty"`
	Key         string   `bson:"key,omitempty"`
	Name        string   `bson:"name,omitempty"`
	Topic       string   `bson:"hash,omitempty"`
	Contracts   []string `bson:"contracts,omitempty"`
	GetLogs     bool     `bson:"get_logs,omitempty"`
	StartBlock  int64    `bson:"start_block,omitempty"`
	EndBlock    int64    `bson:"end_block,omitempty"`
	FilterLogs  string   `bson:"filter_logs,omitempty"`
	FilterValue []string `bson:"filter_value,omitempty"`
}

type DclInfo struct {
	mgm.DefaultModel `bson:",inline"`
	Name             string            `bson:"name,omitempty"`
	Blockchains      []string          `bson:"blockchains,omitempty"`
	Currency         string            `bson:"currency,omitempty"`
	Assets           []DclInfoAsset    `bson:"assets,omitempty"`
	LogTopics        []DclInfoLogTopic `bson:"log_topics,omitempty"`
}

func (m DclInfo) CollectionName() string {
	return "decentraland_info"
}

const (
	LogInfoKeyTransfer1                    = "transfer-asset-1"
	LogInfoKeyTransfer2                    = "transfer-asset-2"
	LogInfoKeyEstatePlus                   = "estate-management-add-land-in-estate"
	LogInfoKeyEstateRemove                 = "estate-management-remove-land-from-estate"
	LogInfoKeyMarketplaceCreateList        = "marketplace-1-list-asset-new"
	LogInfoKeyMarketplaceCancelList        = "marketplace-1-list-asset-cancelled"
	LogInfoKeyMarketplaceSaleList          = "marketplace-1-list-asset-sold"
	LogInfoKeyMarketplaceCreateAsk         = "marketplace-1-ask-asset-new"
	LogInfoKeyMarketplaceCancelAsk         = "marketplace-1-ask-asset-cancelled"
	LogInfoKeyMarketplaceAcceptedAsk       = "marketplace-1-ask-asset-accepted"
	LogInfoKeyMarketplaceV2ListAskAccepted = "marketplace-2-list-or-ask-accepted"
	LogInfoKeyAuction1CreateAsk            = "auction-1-ask-asset-new"
	LogInfoKeyAuction1CancelAsk            = "auction-1-ask-asset-cancelled"
	LogInfoKeyAuction1AcceptedAsk          = "auction-1-ask-asset-accepted"
	LogInfoKeyAuction2AcceptedAsk          = "auction-2-ask-asset-accepted"
)

const (
	LogInfoNameTransfer                     = "Transfer Asset"
	LogInfoNameEstatePlus                   = "Add Land In Estate"
	LogInfoNameEstateRemove                 = "Remove Land From Estate"
	LogInfoNameMarketplaceCreateList        = "DCL Marketplace V1 - New Listing"
	LogInfoNameMarketplaceCancelList        = "DCL Marketplace V1 - Listing Cancelled"
	LogInfoNameMarketplaceSaleList          = "DCL Marketplace V1 - Asset Sold"
	LogInfoNameMarketplaceCreateAsk         = "DCL Marketplace V1 - New Ask"
	LogInfoNameMarketplaceCancelAsk         = "DCL Marketplace V1 - Ask Cancelled"
	LogInfoNameMarketplaceAcceptedAsk       = "DCL Marketplace V1 - Ask Accepted"
	LogInfoNameMarketplaceV2ListAskAccepted = "DCL Marketplace V2 - List or Ask Accepted"
	LogInfoNameAuction1CreateAsk            = "DCL Auction 1 - New Ask"
	LogInfoNameAuction1CancelAsk            = "DCL Auction 1 - Ask Cancelled"
	LogInfoNameAuction1AcceptedAsk          = "DCL Auction 1 - Ask Accepted"
	LogInfoNameAuction2AcceptedAsk          = "DCL Auction 2 - Ask Accepted"
)

const (
	LandContractAddress           = "0xf87e31492faf9a91b02ee0deaad50d51d56d5d4d"
	EstateContractAddress         = "0x959e104e1a4db6317fa58f8295f586e1a978c297"
	MarketplaceV1ContractAddress  = "0x8e5660b4Ab70168b5a6fEeA0e0315cb49c8Cd539"
	BidV1ContractAddress          = "0xE479DfD9664c693b2e2992300930B00bFde08233"
	MarketplaceV2ContractAddress1 = "0x2D6b3508f9Aca32d2550F92b2aDDBa932e73C1ff"
	MarketplaceV2ContractAddress2 = "0x1b67D0e31eeB6B52D8eEEd71D3616C2F5b33b8E7"
	AuctionV1ContractAddress      = "0xb3bca6f5052c7e24726b44da7403b56a8a1b98f8"
	AuctionV2ContractAddress      = "0x54B7a124B44054dA3692dBc56B116a35C6a3e561"
)

const (
	LandKey    = "land"
	LandName   = "Land"
	EstateKey  = "estate"
	EstateName = "Estate"
)

const (
	TransferLogTopic1                   = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	TransferLogTopic2                   = "0xd5c97f2e041b2046be3b4337472f05720760a198f4d7d84980b7155eec7cca6f"
	AddLandInEstateLogTopic             = "0xff0e52667d53255667dc777a00af81038a4646367b0d73d8ee8540ca5b0c9a2e"
	RemoveLandFromEstateLogTopic        = "0x7932eb5ab0d4d4d172776074ee15d13d708465ff5476902ed15a4965434fcab1"
	MarketplaceV1OrderCreatedLogTopic   = "0x84c66c3f7ba4b390e20e8e8233e2a516f3ce34a72749e4f12bd010dfba238039"
	MarketplaceV1OrderSuccessLogTopic   = "0x695ec315e8a642a74d450a4505eeea53df699b47a7378c7d752e97d5b16eb9bb"
	MarketplaceV1OrderCancelledLogTopic = "0x0325426328de5b91ae4ad8462ad4076de4bcaf4551e81556185cacde5a425c6b"
	BidV1BidCreatedTopic                = "0xe45b7709f1eed66d6e28f43b32f2003da0f0c40b92f75a8994370516e048620f"
	BidV1BidCancelledTopic              = "0xc43098075c34b8b92567bd49f08f55e397ebf82dd82072e3eb1be525c4506f5f"
	BidV1BidAcceptedTopic               = "0x4e5ca6c6c06fa8d62f2930830e0d370de70f108bd89213de0b51141775e695bd"
	MarketplaceV2TradedLogTopic         = "0xaaecdfa7e74e704650fcb273f630f42f68974eff42bfffc1732cf30db9e4685b"
	AuctionV1AuctionCreatedLogTopic     = "0x9493ae82b9872af74473effb9d302efba34e0df360a99cc5e577cd3f28e3cab2"
	AuctionV1AuctionCancelledLogTopic   = "0x88bd2ba46f3dc2567144331c35bd4c5ced3d547d8828638a152ddd9591c137a6"
	AuctionV1AuctionSuccessfulLogTopic  = "0xedcc7e1c269bc295dc24e74dc46b129c8449e6b0544af73b57c4201b78d119db"
	AuctionV2BidSuccessfulLogTopic      = "0x640767916efcb255d42962e81c56c9956eaea049461f5c05f5444a8be0b7e896"
)

var DecentralandMtvInfo = &DclInfo{
	Name:        Decentraland,
	Blockchains: []string{EthereumBlockchain},
	Currency:    MANATokenSymbol,
	Assets: []DclInfoAsset{
		{
			Blockchain: EthereumBlockchain,
			Key:        LandKey,
			Name:       LandName,
			Contract:   LandContractAddress,
			AssetType:  AssetTypeLand,
			Attributes: []DclInfoAssetAttribute{
				{
					AttributeName: AssetAttributeNameOwner,
					DataType:      AssetAttributeDataTypeString,
				},
			},
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        EstateKey,
			Name:       EstateName,
			Contract:   EstateContractAddress,
			AssetType:  AssetTypeEstate,
			Attributes: []DclInfoAssetAttribute{
				{
					AttributeName: AssetAttributeNameSize,
					DataType:      AssetAttributeDataTypeInteger,
				},
				{
					AttributeName: AssetAttributeNameOwner,
					DataType:      AssetAttributeDataTypeString,
				},
				{
					AttributeName:  AssetAttributeNameLands,
					DataType:       AssetAttributeDataTypeStringArray,
					DataTypeParams: map[string]any{"separator": "|"},
				},
			},
		},
	},
	LogTopics: []DclInfoLogTopic{
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyTransfer1,
			Name:       LogInfoNameTransfer,
			Topic:      TransferLogTopic1,
			Contracts:  []string{LandContractAddress, EstateContractAddress},
			GetLogs:    true,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyTransfer2,
			Name:       LogInfoNameTransfer,
			Topic:      TransferLogTopic2,
			Contracts:  []string{LandContractAddress, EstateContractAddress},
			GetLogs:    true,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyEstatePlus,
			Name:       LogInfoNameEstatePlus,
			Topic:      AddLandInEstateLogTopic,
			Contracts:  []string{EstateContractAddress},
			GetLogs:    true,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyEstateRemove,
			Name:       LogInfoNameEstateRemove,
			Topic:      RemoveLandFromEstateLogTopic,
			Contracts:  []string{EstateContractAddress},
			GetLogs:    true,
			EndBlock:   0,
		},
		{
			Blockchain:  EthereumBlockchain,
			Key:         LogInfoKeyMarketplaceCreateList,
			Name:        LogInfoNameMarketplaceCreateList,
			Topic:       MarketplaceV1OrderCreatedLogTopic,
			Contracts:   []string{MarketplaceV1ContractAddress},
			GetLogs:     true,
			StartBlock:  0,
			EndBlock:    0,
			FilterLogs:  "data.2",
			FilterValue: []string{LandContractAddress, EstateContractAddress},
		},
		{
			Blockchain:  EthereumBlockchain,
			Key:         LogInfoKeyMarketplaceCancelList,
			Name:        LogInfoNameMarketplaceCancelList,
			Topic:       MarketplaceV1OrderCancelledLogTopic,
			Contracts:   []string{MarketplaceV1ContractAddress},
			GetLogs:     true,
			StartBlock:  0,
			EndBlock:    0,
			FilterLogs:  "data.2",
			FilterValue: []string{LandContractAddress, EstateContractAddress},
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyMarketplaceSaleList,
			Name:       LogInfoNameMarketplaceSaleList,
			Topic:      MarketplaceV1OrderSuccessLogTopic,
			Contracts:  []string{MarketplaceV1ContractAddress},
			GetLogs:    false,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain:  EthereumBlockchain,
			Key:         LogInfoKeyMarketplaceCreateAsk,
			Name:        LogInfoNameMarketplaceCreateAsk,
			Topic:       BidV1BidCreatedTopic,
			Contracts:   []string{BidV1ContractAddress},
			GetLogs:     true,
			StartBlock:  0,
			EndBlock:    0,
			FilterLogs:  "topics.2",
			FilterValue: []string{LandContractAddress, EstateContractAddress},
		},
		{
			Blockchain:  EthereumBlockchain,
			Key:         LogInfoKeyMarketplaceCancelAsk,
			Name:        LogInfoNameMarketplaceCancelAsk,
			Topic:       BidV1BidCancelledTopic,
			Contracts:   []string{BidV1ContractAddress},
			GetLogs:     true,
			StartBlock:  0,
			EndBlock:    0,
			FilterLogs:  "topics.2",
			FilterValue: []string{LandContractAddress, EstateContractAddress},
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyMarketplaceAcceptedAsk,
			Name:       LogInfoNameMarketplaceAcceptedAsk,
			Topic:      BidV1BidAcceptedTopic,
			Contracts:  []string{BidV1ContractAddress},
			GetLogs:    false,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyMarketplaceV2ListAskAccepted,
			Name:       LogInfoNameMarketplaceV2ListAskAccepted,
			Topic:      MarketplaceV2TradedLogTopic,
			Contracts:  []string{MarketplaceV2ContractAddress1, MarketplaceV2ContractAddress2},
			GetLogs:    false,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyAuction1CreateAsk,
			Name:       LogInfoNameAuction1CreateAsk,
			Topic:      AuctionV1AuctionCreatedLogTopic,
			Contracts:  []string{AuctionV1ContractAddress},
			GetLogs:    true,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyAuction1CancelAsk,
			Name:       LogInfoNameAuction1CancelAsk,
			Topic:      AuctionV1AuctionCancelledLogTopic,
			Contracts:  []string{AuctionV1ContractAddress},
			GetLogs:    true,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyAuction1AcceptedAsk,
			Name:       LogInfoNameAuction1AcceptedAsk,
			Topic:      AuctionV1AuctionSuccessfulLogTopic,
			Contracts:  []string{AuctionV1ContractAddress},
			GetLogs:    false,
			StartBlock: 0,
			EndBlock:   0,
		},
		{
			Blockchain: EthereumBlockchain,
			Key:        LogInfoKeyAuction2AcceptedAsk,
			Name:       LogInfoNameAuction2AcceptedAsk,
			Topic:      AuctionV2BidSuccessfulLogTopic,
			Contracts:  []string{AuctionV2ContractAddress},
			GetLogs:    false,
			StartBlock: 0,
			EndBlock:   0,
		},
	},
}

func saveDecentralandInfo(dbInstance *mongo.Database, _dclInfo *DclInfo) (err error) {
	_dclInfo.CreatedAt = time.Now()
	_dclInfo.UpdatedAt = time.Now()
	dbCollection := database.CollectionInstance(dbInstance, &DclInfo{})
	opts := &options.ReplaceOptions{}
	_, err = dbCollection.ReplaceOne(context.Background(), bson.M{"name": Decentraland}, _dclInfo, opts.SetUpsert(true))
	return
}

func getDecentralandInfo(dbInstance *mongo.Database) (dclInfo *DclInfo, err error) {
	dbCollection := database.CollectionInstance(dbInstance, &DclInfo{})
	dclInfo = &DclInfo{}
	err = dbCollection.FirstWithCtx(context.Background(), bson.M{"name": Decentraland}, dclInfo)
	return
}

func GetDecentralandInfo(dbInstance *mongo.Database) (dclInfo *DclInfo, err error) {
	dclInfo, err = getDecentralandInfo(dbInstance)
	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		err = saveDecentralandInfo(dbInstance, DecentralandMtvInfo)
		if err != nil {
			return nil, err
		}
		dclInfo, err = getDecentralandInfo(dbInstance)
	}
	return
}

func SaveLastBlockNumber(dbInstance *mongo.Database, infoLogTopics map[string]*DclInfoLogTopic) (err error) {
	dbCollection := database.CollectionInstance(dbInstance, &DclInfo{})
	for _, infoLogTopic := range infoLogTopics {
		for i, _infoLogTopic := range DecentralandMtvInfo.LogTopics {
			if infoLogTopic.Key == _infoLogTopic.Key {
				DecentralandMtvInfo.LogTopics[i].StartBlock = infoLogTopic.StartBlock
				break
			}
		}
	}
	DecentralandMtvInfo.UpdatedAt = time.Now()
	opts := &options.ReplaceOptions{}
	_, err = dbCollection.ReplaceOne(context.Background(), bson.M{"name": Decentraland}, DecentralandMtvInfo, opts.SetUpsert(true))
	return
}
