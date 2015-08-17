package otr3

import (
	"bytes"
	"math/big"
)

type otrVersion interface {
	protocolVersion() uint16
	parameterLength() int
	isGroupElement(n *big.Int) bool
	isFragmented(data []byte) bool
	parseFragmentPrefix(c *Conversation, data []byte) (rest []byte, ignore bool, ok bool)
	fragmentPrefix(n, total int, itags uint32, itagr uint32) []byte
	whitespaceTag() []byte
	messageHeader(c *Conversation, msgType byte) ([]byte, error)
	parseMessageHeader(c *Conversation, msg []byte) ([]byte, []byte, error)
}

func newOtrVersion(v uint16, p policies) (version otrVersion, err error) {
	toCheck := policy(0)
	switch v {
	case 2:
		version = otrV2{}
		toCheck = allowV2
	case 3:
		version = otrV3{}
		toCheck = allowV3
	default:
		return nil, errUnsupportedOTRVersion
	}
	if !p.has(toCheck) {
		return nil, errInvalidVersion
	}
	return
}

func versionFromFragment(fragment []byte) uint16 {
	var messageVersion uint16

	switch {
	case bytes.HasPrefix(fragment, otrv3FragmentationPrefix):
		messageVersion = 3
	case bytes.HasPrefix(fragment, otrv2FragmentationPrefix):
		messageVersion = 2
	}

	return messageVersion
}

func (c *Conversation) checkVersion(message []byte) (err error) {
	_, messageVersion, ok := extractShort(message)
	if !ok {
		return errInvalidOTRMessage
	}

	if c.version == nil {
		if c.version, err = newOtrVersion(messageVersion, c.Policies); err != nil {
			return err
		}
	}

	if c.version.protocolVersion() != messageVersion {
		return errWrongProtocolVersion
	}

	return nil
}

func (c *Conversation) resolveVersion(versions int) error {
	switch {
	case c.Policies.has(allowV3) && versions&(1<<3) > 0:
		c.version = otrV3{}
	case c.Policies.has(allowV2) && versions&(1<<2) > 0:
		c.version = otrV2{}
	default:
		return errInvalidVersion
	}

	return nil
}
