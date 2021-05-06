//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/gomodule/redigo/redis"
)

const (
	TransmissionCollection                 = "sn|trans"
	TransmissionCollectionStatus           = TransmissionCollection + DBKeySeparator + v2.Status
	TransmissionCollectionSubscriptionName = TransmissionCollection + DBKeySeparator + v2.Subscription + DBKeySeparator + v2.Name
	TransmissionCollectionNotificationId   = TransmissionCollection + DBKeySeparator + v2.Notification + DBKeySeparator + v2.Id
)

// notificationStoredKey return the transmission's stored key which combines the collection name and object id
func transmissionStoredKey(id string) string {
	return CreateKey(TransmissionCollection, id)
}

// transmissionById query transmission by id from DB
func transmissionById(conn redis.Conn, id string) (trans models.Transmission, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, transmissionStoredKey(id), &trans)
	if edgexErr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	return
}

// sendAddTransmissionCmd sends redis command for adding transmission
func sendAddTransmissionCmd(conn redis.Conn, storedKey string, trans models.Transmission) errors.EdgeX {
	m, err := json.Marshal(trans)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal transmission for Redis persistence", err)
	}
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, TransmissionCollection, trans.Modified, storedKey)
	_ = conn.Send(ZADD, CreateKey(TransmissionCollectionStatus, string(trans.Status)), trans.Modified, storedKey)
	_ = conn.Send(ZADD, CreateKey(TransmissionCollectionSubscriptionName, trans.SubscriptionName), trans.Modified, storedKey)
	_ = conn.Send(ZADD, CreateKey(TransmissionCollectionNotificationId, trans.NotificationId), trans.Modified, storedKey)
	return nil
}

// addTransmission adds a new transmission into DB
func addTransmission(conn redis.Conn, trans models.Transmission) (models.Transmission, errors.EdgeX) {
	exists, edgeXerr := objectIdExists(conn, transmissionStoredKey(trans.Id))
	if edgeXerr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return trans, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("transmission id %s already exists", trans.Id), edgeXerr)
	}

	ts := common.MakeTimestamp()
	if trans.Created == 0 {
		trans.Created = ts
	}
	trans.Modified = ts

	storedKey := transmissionStoredKey(trans.Id)
	_ = conn.Send(MULTI)
	edgeXerr = sendAddTransmissionCmd(conn, storedKey, trans)
	if edgeXerr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "transmission creation failed", err)
	}

	return trans, edgeXerr
}

// sendDeleteTransmissionCmd sends redis command to delete a transmission
func sendDeleteTransmissionCmd(conn redis.Conn, storedKey string, trans models.Transmission) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, TransmissionCollection, storedKey)
	_ = conn.Send(ZREM, CreateKey(TransmissionCollectionStatus, string(trans.Status)), storedKey)
	_ = conn.Send(ZREM, CreateKey(TransmissionCollectionSubscriptionName, trans.SubscriptionName), storedKey)
	_ = conn.Send(ZREM, CreateKey(TransmissionCollectionNotificationId, trans.NotificationId), storedKey)
}

// updateTransmission updates a transmission
func updateTransmission(conn redis.Conn, trans models.Transmission) errors.EdgeX {
	oldTransmission, edgeXerr := transmissionById(conn, trans.Id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	trans.Modified = common.MakeTimestamp()
	storedKey := transmissionStoredKey(trans.Id)

	_ = conn.Send(MULTI)
	sendDeleteTransmissionCmd(conn, storedKey, oldTransmission)
	edgeXerr = sendAddTransmissionCmd(conn, storedKey, trans)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "transmission update failed", err)
	}
	return nil
}