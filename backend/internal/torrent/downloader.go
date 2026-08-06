package torrent

import (
	"fmt"
	"time"
)

const MaxBlockSize = 16 * 1024
const MaxBacklog = 5

func AttemptDownloadPiece(c *Client, pw *PieceWork) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if pw == nil {
		return nil, fmt.Errorf("piece work is nil")
	}

	if err := c.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set piece deadline: %w", err)
	}
	defer func() { _ = c.SetDeadline(time.Time{}) }()

	if err := c.SendInterested(); err != nil {
		return nil, fmt.Errorf("failed to send interested: %w", err)
	}

	requested := 0
	backlog := 0
	for !pw.Complete() {
		msg, err := c.Read()
		if err != nil {
			return nil, err
		}
		if msg == nil {
			continue
		}

		switch msg.ID {
		case MsgChoke:
			c.Choked = true
			backlog = 0
			requested = pw.Downloaded
		case MsgUnchoke:
			c.Choked = false
			if err := fillRequestPipeline(c, pw, &requested, &backlog); err != nil {
				return nil, err
			}
		case MsgPiece:
			piece, err := ParsePiece(msg)
			if err != nil {
				return nil, err
			}
			if piece.Index != pw.Index {
				continue
			}
			if err := pw.WriteBlock(piece.Begin, piece.Block); err != nil {
				return nil, fmt.Errorf("failed to store block: %w", err)
			}
			if backlog > 0 {
				backlog--
			}
			if !c.Choked {
				if err := fillRequestPipeline(c, pw, &requested, &backlog); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := pw.Verify(); err != nil {
		return nil, err
	}
	return append([]byte(nil), pw.Buffer...), nil
}

func fillRequestPipeline(c *Client, pw *PieceWork, requested *int, backlog *int) error {
	for *backlog < MaxBacklog && *requested < pw.Length {
		blockSize := MaxBlockSize
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
