package wasmtesting

import (
	"context"
	"errors"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v3/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm/types"
)

type MockChannelKeeper struct {
	GetChannelFn                   func(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSendFn          func(ctx sdk.Context, portID, channelID string) (uint64, bool)
	ChanCloseInitFn                func(ctx sdk.Context, portID, channelID string) error
	GetAllChannelsFn               func(ctx sdk.Context) []channeltypes.IdentifiedChannel
	SetChannelFn                   func(ctx sdk.Context, portID, channelID string, channel channeltypes.Channel)
	GetAllChannelsWithPortPrefixFn func(ctx sdk.Context, portPrefix string) []channeltypes.IdentifiedChannel
}

func (m *MockChannelKeeper) GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool) {
	if m.GetChannelFn == nil {
		panic("not supposed to be called!")
	}
	return m.GetChannelFn(ctx, srcPort, srcChan)
}

func (m *MockChannelKeeper) GetAllChannels(ctx sdk.Context) []channeltypes.IdentifiedChannel {
	if m.GetAllChannelsFn == nil {
		panic("not supposed to be called!")
	}
	return m.GetAllChannelsFn(ctx)
}

func (m *MockChannelKeeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	if m.GetNextSequenceSendFn == nil {
		panic("not supposed to be called!")
	}
	return m.GetNextSequenceSendFn(ctx, portID, channelID)
}

func (m *MockChannelKeeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	if m.ChanCloseInitFn == nil {
		panic("not supposed to be called!")
	}
	return m.ChanCloseInitFn(ctx, portID, channelID)
}

func (m *MockChannelKeeper) GetAllChannelsWithPortPrefix(ctx sdk.Context, portPrefix string) []channeltypes.IdentifiedChannel {
	if m.GetAllChannelsWithPortPrefixFn == nil {
		panic("not expected to be called")
	}
	return m.GetAllChannelsWithPortPrefixFn(ctx, portPrefix)
}

func (m *MockChannelKeeper) SetChannel(ctx sdk.Context, portID, channelID string, channel channeltypes.Channel) {
	if m.GetChannelFn == nil {
		panic("not supposed to be called!")
	}
	m.SetChannelFn(ctx, portID, channelID, channel)
}

var _ types.ICS4Wrapper = &MockICS4Wrapper{}

type MockICS4Wrapper struct {
	SendPacketFn           func(ctx sdk.Context, sourcePort, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error)
	WriteAcknowledgementFn func(ctx sdk.Context, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error
}

func (m *MockICS4Wrapper) SendPacket(ctx sdk.Context, sourcePort, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error) {
	if m.SendPacketFn == nil {
		panic("not supposed to be called!")
	}
	return m.SendPacketFn(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

func (m *MockICS4Wrapper) WriteAcknowledgement(
	ctx sdk.Context,
	packet ibcexported.PacketI,
	acknowledgement ibcexported.Acknowledgement,
) error {
	if m.WriteAcknowledgementFn == nil {
		panic("not supposed to be called!")
	}
	return m.WriteAcknowledgementFn(ctx, packet, acknowledgement)
}

func MockChannelKeeperIterator(s []channeltypes.IdentifiedChannel) func(ctx sdk.Context, cb func(channeltypes.IdentifiedChannel) bool) {
	return func(ctx sdk.Context, cb func(channeltypes.IdentifiedChannel) bool) {
		for _, channel := range s {
			stop := cb(channel)
			if stop {
				break
			}
		}
	}
}

var _ types.ICS20TransferPortSource = &MockIBCTransferKeeper{}

type MockIBCTransferKeeper struct {
	GetPortFn func(ctx sdk.Context) string
}

func (m MockIBCTransferKeeper) GetPort(ctx sdk.Context) string {
	if m.GetPortFn == nil {
		panic("not expected to be called")
	}
	return m.GetPortFn(ctx)
}

var _ types.IBCContractKeeper = &IBCContractKeeperMock{}

type IBCContractKeeperMock struct {
	types.IBCContractKeeper
	OnRecvPacketFn func(ctx sdk.Context, contractAddr sdk.AccAddress, msg wasmvmtypes.IBCPacketReceiveMsg) (ibcexported.Acknowledgement, error)

	packets map[string]channeltypes.Packet
}

func (m *IBCContractKeeperMock) OnRecvPacket(ctx sdk.Context, contractAddr sdk.AccAddress, msg wasmvmtypes.IBCPacketReceiveMsg) (ibcexported.Acknowledgement, error) {
	if m.OnRecvPacketFn == nil {
		panic("not expected to be called")
	}
	return m.OnRecvPacketFn(ctx, contractAddr, msg)
}

func (m *IBCContractKeeperMock) LoadAsyncAckPacket(ctx context.Context, portID, channelID string, sequence uint64) (channeltypes.Packet, error) {
	if m.packets == nil {
		m.packets = make(map[string]channeltypes.Packet)
	}
	key := portID + fmt.Sprint(len(channelID)) + channelID
	packet, ok := m.packets[key]
	if !ok {
		return channeltypes.Packet{}, errors.New("packet not found")
	}
	return packet, nil
}

func (m *IBCContractKeeperMock) StoreAsyncAckPacket(ctx context.Context, packet channeltypes.Packet) error {
	if m.packets == nil {
		m.packets = make(map[string]channeltypes.Packet)
	}
	key := packet.DestinationPort + fmt.Sprint(len(packet.DestinationChannel)) + packet.DestinationChannel
	m.packets[key] = packet
	return nil
}

func (m *IBCContractKeeperMock) DeleteAsyncAckPacket(ctx context.Context, portID, channelID string, sequence uint64) {
	key := portID + fmt.Sprint(len(channelID)) + channelID
	delete(m.packets, key)
}
