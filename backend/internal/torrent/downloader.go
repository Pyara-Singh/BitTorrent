package torrent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// Standard BitTorrent block size (16 KB).
const MaxBlockSize = 16 * 1024

// Maximum number of outstanding requests allowed at once.
const MaxBacklog = 5

// AttemptDownloadPiece downloads one complete piece from a single peer.
//
// Flow:
//
//	Interested
//	      ↓
//	Wait for Unchoke
//	      ↓
//	Send block requests
//	      ↓
//	Receive block data
//	      ↓
//	Store blocks in PieceWork
//	      ↓
//	Request remaining blocks
//	      ↓
//	Verify SHA-1
//	      ↓
//	Return completed piece
func AttemptDownloadPiece(c *Client, pw *PieceWork) ([]byte, error) {

	// Prevent a peer from blocking forever while downloading.
	if err := c.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return nil, err
	}
	defer c.SetDeadline(time.Time{})

	// Tell the peer that we want to download pieces.
	if err := c.SendInterested(); err != nil {
		return nil, fmt.Errorf("failed to send interested: %w", err)
	}

	// Next byte offset that should be requested.
	requested := 0

	// Number of requests currently waiting for a response.
	backlog := 0

	// Continue until the entire piece has been assembled.
	for pw.Downloaded < pw.Length {

		msg, err := c.Read()
		if err != nil {
			return nil, fmt.Errorf("read peer message failed: %w", err)
		}

		// Nil represents a KeepAlive message.
		if msg == nil {
			continue
		}

		switch msg.ID {

		case MsgChoke:

			// Peer stopped serving us.
			c.Choked = true

			// Outstanding requests are assumed lost.
			backlog = 0
			requested = pw.Downloaded

		case MsgUnchoke:

			// Peer is ready to upload blocks.
			c.Choked = false

			// Fill the request pipeline.
			if err := fillRequestPipeline(c, pw, &requested, &backlog); err != nil {
				return nil, err
			}

		case MsgPiece:

			// Piece payload layout:
			//
			//	0-3  -> Piece Index
			//	4-7  -> Block Offset
			//	8+   -> Block Data
			if len(msg.Payload) < 8 {
				return nil, errors.New("malformed piece payload")
			}

			index := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
			begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
			block := msg.Payload[8:]

			// Ignore blocks that belong to another piece.
			if index != pw.Index {
				continue
			}

			// Copy the received block into the piece buffer.
			if err := pw.WriteBlock(begin, block); err != nil {
				return nil, fmt.Errorf("failed to store block: %w", err)
			}

			backlog--

			// Keep the pipeline full while the peer is willing to upload.
			if !c.Choked {
				if err := fillRequestPipeline(c, pw, &requested, &backlog); err != nil {
					return nil, err
				}
			}

		default:

			// Ignore messages not needed during piece download.
			continue
		}
	}

	// Verify the completed piece before accepting it.
	if err := pw.Verify(); err != nil {
		return nil, err
	}

	return pw.Buffer, nil
}

// fillRequestPipeline keeps multiple block requests in flight.
//
// Instead of:
//
//	Request
//	↓
//	Wait
//	↓
//	Request
//
// we do:
//
//	Request
//	Request
//	Request
//	Request
//	Request
//
// which greatly improves download speed.
func fillRequestPipeline(c *Client, pw *PieceWork, requested *int, backlog *int) error {

	for *backlog < MaxBacklog && *requested < pw.Length {

		blockSize := MaxBlockSize

		// Last block of a piece may be smaller than 16 KB.
		if remaining := pw.Length - *requested; remaining < blockSize {
			blockSize = remaining
		}

		if err := c.SendRequest(pw.Index, *requested, blockSize); err != nil {
			return fmt.Errorf("failed to send block request: %w", err)
		}

		*backlog++
		*requested += blockSize
	}

	return nil
}
