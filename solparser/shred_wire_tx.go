package solparser

import (
	solana "github.com/gagliardetto/solana-go"
)

// DexEventsFromShredTransactionWire 从 Shred / gRPC 负载中的**线格式交易字节**解析外层编译指令并调用
// ParseInstructionUnified（与 sol-parser-sdk-ts `dexEventsFromShredWasmTx` 一致）。
//
// 限制：仅使用消息中的**静态**账户表。V0 交易若指令账户索引指向 ALT 加载地址且尚未展开，会跳过该条指令（与 TS 在 oob 时 continue 一致）。
func DexEventsFromShredTransactionWire(
	raw []byte,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
	filter EventTypeFilter,
) []DexEvent {
	tx, err := solana.TransactionFromBytes(raw)
	if err != nil {
		return nil
	}
	msg := tx.Message
	keys := msg.AccountKeys
	if len(keys) == 0 {
		return nil
	}

	var out []DexEvent
	for _, ix := range msg.Instructions {
		if int(ix.ProgramIDIndex) >= len(keys) {
			continue
		}
		pid := keys[ix.ProgramIDIndex].String()

		accStrs := make([]string, 0, len(ix.Accounts))
		ok := true
		for _, ai := range ix.Accounts {
			if int(ai) >= len(keys) {
				ok = false
				break
			}
			accStrs = append(accStrs, keys[ai].String())
		}
		if !ok {
			continue
		}

		data := []byte(ix.Data)
		ev := ParseInstructionUnified(data, accStrs, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, pid)
		if ev.Type != "" {
			out = append(out, ev)
		}
	}
	return out
}
