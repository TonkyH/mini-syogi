package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// 駒の種類
type PieceType int

const (
	Empty          PieceType = iota
	King                     // 玉
	Gold                     // 金
	Silver                   // 銀
	Bishop                   // 角
	Rook                     // 飛
	Pawn                     // 歩
	PromotedSilver           // 成銀
	PromotedBishop           // 成角（馬）
	PromotedRook             // 成飛（龍）
	PromotedPawn             // と金
)

// プレイヤー
type Player int

const (
	None Player = iota
	First
	Second
)

// 駒
type Piece struct {
	Type  PieceType
	Owner Player
}

// 盤面
type Board struct {
	Cells       [5][5]Piece
	FirstHand   []PieceType // 先手の持ち駒
	SecondHand  []PieceType // 後手の持ち駒
	CurrentTurn Player
}

// 移動
type Move struct {
	FromRow, FromCol int
	ToRow, ToCol     int
	IsDrop           bool
	DropPiece        PieceType
	Promote          bool
}

// ゲーム初期化
func NewBoard() *Board {
	b := &Board{
		FirstHand:   []PieceType{},
		SecondHand:  []PieceType{},
		CurrentTurn: First,
	}

	// 初期配置（5五将棋の標準配置）
	// 後手（上側）
	b.Cells[0][0] = Piece{Rook, Second}
	b.Cells[0][1] = Piece{Bishop, Second}
	b.Cells[0][2] = Piece{Silver, Second}
	b.Cells[0][3] = Piece{Gold, Second}
	b.Cells[0][4] = Piece{King, Second}
	b.Cells[1][4] = Piece{Pawn, Second}

	// 先手（下側）
	b.Cells[4][4] = Piece{Rook, First}
	b.Cells[4][3] = Piece{Bishop, First}
	b.Cells[4][2] = Piece{Silver, First}
	b.Cells[4][1] = Piece{Gold, First}
	b.Cells[4][0] = Piece{King, First}
	b.Cells[3][0] = Piece{Pawn, First}

	return b
}

// 駒の文字表現
func (p Piece) String() string {
	if p.Owner == None {
		return " ． "
	}

	var symbol string
	switch p.Type {
	case King:
		symbol = "玉"
	case Gold:
		symbol = "金"
	case Silver:
		symbol = "銀"
	case Bishop:
		symbol = "角"
	case Rook:
		symbol = "飛"
	case Pawn:
		symbol = "歩"
	case PromotedSilver:
		symbol = "全"
	case PromotedBishop:
		symbol = "馬"
	case PromotedRook:
		symbol = "龍"
	case PromotedPawn:
		symbol = "と"
	}

	if p.Owner == First {
		return " " + symbol + " "
	} else {
		return "v" + symbol + " "
	}
}

// 盤面表示
func (b *Board) Display() {
	fmt.Println("\n  １ ２ ３ ４ ５")
	fmt.Println("┌─────────────┐")
	for i := 0; i < 5; i++ {
		fmt.Printf("│")
		for j := 0; j < 5; j++ {
			fmt.Printf("%s", b.Cells[i][j])
		}
		fmt.Printf("│%s\n", []string{"一", "二", "三", "四", "五"}[i])
	}
	fmt.Println("└─────────────┘")

	// 持ち駒表示
	fmt.Printf("先手持ち駒: ")
	b.displayHand(b.FirstHand)
	fmt.Printf("後手持ち駒: ")
	b.displayHand(b.SecondHand)
}

func (b *Board) displayHand(hand []PieceType) {
	if len(hand) == 0 {
		fmt.Println("なし")
		return
	}
	counts := make(map[PieceType]int)
	for _, p := range hand {
		counts[p]++
	}
	for pType, count := range counts {
		piece := Piece{Type: pType, Owner: First}
		fmt.Printf("%s×%d ", strings.TrimSpace(piece.String()), count)
	}
	fmt.Println()
}

// 移動可能な位置を取得
func (b *Board) GetPossibleMoves(row, col int) []Move {
	piece := b.Cells[row][col]
	if piece.Owner == None || piece.Owner != b.CurrentTurn {
		return []Move{}
	}

	moves := []Move{}

	switch piece.Type {
	case King:
		// 8方向に1マス
		dirs := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
		for _, d := range dirs {
			nr, nc := row+d[0], col+d[1]
			if b.isValidMove(row, col, nr, nc) {
				moves = append(moves, Move{row, col, nr, nc, false, Empty, false})
			}
		}

	case Gold, PromotedSilver, PromotedPawn:
		// 金の動き
		dirs := b.getGoldMoves(piece.Owner)
		for _, d := range dirs {
			nr, nc := row+d[0], col+d[1]
			if b.isValidMove(row, col, nr, nc) {
				moves = append(moves, Move{row, col, nr, nc, false, Empty, false})
			}
		}

	case Silver:
		// 銀の動き
		dirs := b.getSilverMoves(piece.Owner)
		for _, d := range dirs {
			nr, nc := row+d[0], col+d[1]
			if b.isValidMove(row, col, nr, nc) {
				move := Move{row, col, nr, nc, false, Empty, false}
				// 成りの判定
				if b.canPromote(piece.Owner, nr) {
					moves = append(moves, Move{row, col, nr, nc, false, Empty, true})
				}
				moves = append(moves, move)
			}
		}

	case Bishop, PromotedBishop:
		// 斜め方向
		dirs := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
		for _, d := range dirs {
			for i := 1; i < 5; i++ {
				nr, nc := row+d[0]*i, col+d[1]*i
				if !b.isInBoard(nr, nc) {
					break
				}
				if b.Cells[nr][nc].Owner == piece.Owner {
					break
				}
				move := Move{row, col, nr, nc, false, Empty, false}
				if piece.Type == Bishop && b.canPromote(piece.Owner, nr) {
					moves = append(moves, Move{row, col, nr, nc, false, Empty, true})
				}
				moves = append(moves, move)
				if b.Cells[nr][nc].Owner != None {
					break
				}
			}
		}
		// 馬の場合は1マス直進も可能
		if piece.Type == PromotedBishop {
			dirs = [][2]int{{-1, 0}, {0, -1}, {0, 1}, {1, 0}}
			for _, d := range dirs {
				nr, nc := row+d[0], col+d[1]
				if b.isValidMove(row, col, nr, nc) {
					moves = append(moves, Move{row, col, nr, nc, false, Empty, false})
				}
			}
		}

	case Rook, PromotedRook:
		// 直線方向
		dirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
		for _, d := range dirs {
			for i := 1; i < 5; i++ {
				nr, nc := row+d[0]*i, col+d[1]*i
				if !b.isInBoard(nr, nc) {
					break
				}
				if b.Cells[nr][nc].Owner == piece.Owner {
					break
				}
				move := Move{row, col, nr, nc, false, Empty, false}
				if piece.Type == Rook && b.canPromote(piece.Owner, nr) {
					moves = append(moves, Move{row, col, nr, nc, false, Empty, true})
				}
				moves = append(moves, move)
				if b.Cells[nr][nc].Owner != None {
					break
				}
			}
		}
		// 龍の場合は斜め1マスも可能
		if piece.Type == PromotedRook {
			dirs = [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
			for _, d := range dirs {
				nr, nc := row+d[0], col+d[1]
				if b.isValidMove(row, col, nr, nc) {
					moves = append(moves, Move{row, col, nr, nc, false, Empty, false})
				}
			}
		}

	case Pawn:
		// 前進のみ
		dir := 1
		if piece.Owner == Second {
			dir = 1
		} else {
			dir = -1
		}
		nr := row + dir
		if b.isValidMove(row, col, nr, col) {
			move := Move{row, col, nr, col, false, Empty, false}
			if b.canPromote(piece.Owner, nr) {
				moves = append(moves, Move{row, col, nr, col, false, Empty, true})
			}
			moves = append(moves, move)
		}
	}

	return moves
}

// 持ち駒を打つ手を取得
func (b *Board) GetDropMoves() []Move {
	moves := []Move{}
	hand := b.FirstHand
	if b.CurrentTurn == Second {
		hand = b.SecondHand
	}

	// 重複を除く
	uniquePieces := make(map[PieceType]bool)
	for _, p := range hand {
		uniquePieces[p] = true
	}

	for pType := range uniquePieces {
		for r := 0; r < 5; r++ {
			for c := 0; c < 5; c++ {
				if b.Cells[r][c].Owner == None {
					// 歩の二歩チェック
					if pType == Pawn && b.hasPawnInColumn(c, b.CurrentTurn) {
						continue
					}
					// 行き所のない駒チェック
					if pType == Pawn {
						if (b.CurrentTurn == First && r == 0) || (b.CurrentTurn == Second && r == 4) {
							continue
						}
					}
					moves = append(moves, Move{-1, -1, r, c, true, pType, false})
				}
			}
		}
	}

	return moves
}

// 全ての合法手を取得
func (b *Board) GetAllLegalMoves() []Move {
	moves := []Move{}

	// 盤上の駒の移動
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			if b.Cells[r][c].Owner == b.CurrentTurn {
				moves = append(moves, b.GetPossibleMoves(r, c)...)
			}
		}
	}

	// 持ち駒を打つ
	moves = append(moves, b.GetDropMoves()...)

	return moves
}

// 移動実行
func (b *Board) MakeMove(move Move) bool {
	if move.IsDrop {
		// 持ち駒を打つ
		b.Cells[move.ToRow][move.ToCol] = Piece{move.DropPiece, b.CurrentTurn}
		// 持ち駒から削除
		hand := &b.FirstHand
		if b.CurrentTurn == Second {
			hand = &b.SecondHand
		}
		for i, p := range *hand {
			if p == move.DropPiece {
				*hand = append((*hand)[:i], (*hand)[i+1:]...)
				break
			}
		}
	} else {
		// 通常の移動
		piece := b.Cells[move.FromRow][move.FromCol]
		captured := b.Cells[move.ToRow][move.ToCol]

		// 駒を取る
		if captured.Owner != None {
			capturedType := captured.Type
			// 成り駒は元に戻す
			switch capturedType {
			case PromotedSilver:
				capturedType = Silver
			case PromotedBishop:
				capturedType = Bishop
			case PromotedRook:
				capturedType = Rook
			case PromotedPawn:
				capturedType = Pawn
			}

			if b.CurrentTurn == First {
				b.FirstHand = append(b.FirstHand, capturedType)
			} else {
				b.SecondHand = append(b.SecondHand, capturedType)
			}
		}

		// 成り
		if move.Promote {
			switch piece.Type {
			case Silver:
				piece.Type = PromotedSilver
			case Bishop:
				piece.Type = PromotedBishop
			case Rook:
				piece.Type = PromotedRook
			case Pawn:
				piece.Type = PromotedPawn
			}
		}

		b.Cells[move.ToRow][move.ToCol] = piece
		b.Cells[move.FromRow][move.FromCol] = Piece{Empty, None}
	}

	// ターン交代
	if b.CurrentTurn == First {
		b.CurrentTurn = Second
	} else {
		b.CurrentTurn = First
	}

	return true
}

// ヘルパー関数
func (b *Board) isInBoard(row, col int) bool {
	return row >= 0 && row < 5 && col >= 0 && col < 5
}

func (b *Board) isValidMove(fromRow, fromCol, toRow, toCol int) bool {
	if !b.isInBoard(toRow, toCol) {
		return false
	}
	target := b.Cells[toRow][toCol]
	piece := b.Cells[fromRow][fromCol]
	return target.Owner != piece.Owner
}

func (b *Board) canPromote(player Player, row int) bool {
	if player == First {
		return row <= 0
	}
	return row >= 4
}

func (b *Board) getGoldMoves(player Player) [][2]int {
	if player == First {
		return [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, 0}}
	}
	return [][2]int{{1, -1}, {1, 0}, {1, 1}, {0, -1}, {0, 1}, {-1, 0}}
}

func (b *Board) getSilverMoves(player Player) [][2]int {
	if player == First {
		return [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {1, -1}, {1, 1}}
	}
	return [][2]int{{1, -1}, {1, 0}, {1, 1}, {-1, -1}, {-1, 1}}
}

func (b *Board) hasPawnInColumn(col int, player Player) bool {
	for r := 0; r < 5; r++ {
		if b.Cells[r][col].Owner == player && b.Cells[r][col].Type == Pawn {
			return true
		}
	}
	return false
}

// 勝敗判定
func (b *Board) IsGameOver() (bool, Player) {
	// 玉が取られたかチェック
	firstKing, secondKing := false, false
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			if b.Cells[r][c].Type == King {
				if b.Cells[r][c].Owner == First {
					firstKing = true
				} else if b.Cells[r][c].Owner == Second {
					secondKing = true
				}
			}
		}
	}

	if !firstKing {
		return true, Second
	}
	if !secondKing {
		return true, First
	}

	// TODO: 詰みチェック（簡易版では省略）

	return false, None
}

// AI: 評価関数
func (b *Board) Evaluate() int {
	score := 0
	pieceValues := map[PieceType]int{
		King:           10000,
		Gold:           600,
		Silver:         500,
		Bishop:         800,
		Rook:           900,
		Pawn:           100,
		PromotedSilver: 600,
		PromotedBishop: 1000,
		PromotedRook:   1100,
		PromotedPawn:   600,
	}

	// 盤上の駒
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			piece := b.Cells[r][c]
			if piece.Owner == First {
				score += pieceValues[piece.Type]
			} else if piece.Owner == Second {
				score -= pieceValues[piece.Type]
			}
		}
	}

	// 持ち駒
	for _, p := range b.FirstHand {
		score += pieceValues[p] * 8 / 10
	}
	for _, p := range b.SecondHand {
		score -= pieceValues[p] * 8 / 10
	}

	return score
}

// AI: ミニマックス法
func (b *Board) Minimax(depth int, alpha, beta int, maximizing bool) (int, *Move) {
	if depth == 0 {
		return b.Evaluate(), nil
	}

	gameOver, _ := b.IsGameOver()
	if gameOver {
		return b.Evaluate(), nil
	}

	moves := b.GetAllLegalMoves()
	if len(moves) == 0 {
		return b.Evaluate(), nil
	}

	var bestMove *Move
	if maximizing {
		maxEval := -999999
		for _, move := range moves {
			// コピーを作成
			newBoard := *b
			newBoard.FirstHand = append([]PieceType{}, b.FirstHand...)
			newBoard.SecondHand = append([]PieceType{}, b.SecondHand...)

			newBoard.MakeMove(move)
			eval, _ := newBoard.Minimax(depth-1, alpha, beta, false)

			if eval > maxEval {
				maxEval = eval
				moveCopy := move
				bestMove = &moveCopy
			}

			alpha = max(alpha, eval)
			if beta <= alpha {
				break
			}
		}
		return maxEval, bestMove
	} else {
		minEval := 999999
		for _, move := range moves {
			// コピーを作成
			newBoard := *b
			newBoard.FirstHand = append([]PieceType{}, b.FirstHand...)
			newBoard.SecondHand = append([]PieceType{}, b.SecondHand...)

			newBoard.MakeMove(move)
			eval, _ := newBoard.Minimax(depth-1, alpha, beta, true)

			if eval < minEval {
				minEval = eval
				moveCopy := move
				bestMove = &moveCopy
			}

			beta = min(beta, eval)
			if beta <= alpha {
				break
			}
		}
		return minEval, bestMove
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AIの手を取得
func (b *Board) GetAIMove() *Move {
	depth := 3 // 探索深度
	_, move := b.Minimax(depth, -999999, 999999, b.CurrentTurn == First)
	return move
}

// メインゲームループ
func main() {
	rand.Seed(time.Now().UnixNano())
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== ミニ将棋（5五将棋）===")
	fmt.Println("1: 先手（人間） vs 後手（AI）")
	fmt.Println("2: 先手（AI） vs 後手（人間）")
	fmt.Print("選択してください: ")

	scanner.Scan()
	mode, _ := strconv.Atoi(scanner.Text())

	board := NewBoard()
	aiPlayer := Second
	if mode == 2 {
		aiPlayer = First
	}

	for {
		board.Display()

		gameOver, winner := board.IsGameOver()
		if gameOver {
			if winner == First {
				fmt.Println("\n先手の勝ちです！")
			} else {
				fmt.Println("\n後手の勝ちです！")
			}
			break
		}

		if board.CurrentTurn == First {
			fmt.Println("\n先手の番です")
		} else {
			fmt.Println("\n後手の番です")
		}

		var move *Move

		if board.CurrentTurn == aiPlayer {
			fmt.Println("AIが考えています...")
			move = board.GetAIMove()
			if move != nil {
				if move.IsDrop {
					piece := Piece{Type: move.DropPiece, Owner: First}
					fmt.Printf("AI: %sを%d%sに打つ\n",
						strings.TrimSpace(piece.String()),
						move.ToCol+1,
						[]string{"一", "二", "三", "四", "五"}[move.ToRow])
				} else {
					fmt.Printf("AI: %d%sから%d%sへ",
						move.FromCol+1,
						[]string{"一", "二", "三", "四", "五"}[move.FromRow],
						move.ToCol+1,
						[]string{"一", "二", "三", "四", "五"}[move.ToRow])
					if move.Promote {
						fmt.Print("（成）")
					}
					fmt.Println()
				}
			}
		} else {
			// 人間の入力
			fmt.Println("移動: 5133 のように入力（51から33へ）")
			fmt.Println("持ち駒: p53 のように入力（p=歩,s=銀,g=金,b=角,r=飛を53に打つ）")
			fmt.Print("入力: ")

			scanner.Scan()
			input := scanner.Text()

			move = parseInput(input, board)
			if move == nil {
				fmt.Println("無効な入力です")
				continue
			}

			// 合法手チェック
			legalMoves := board.GetAllLegalMoves()
			found := false
			for _, lm := range legalMoves {
				if movesEqual(move, &lm) {
					move = &lm
					found = true
					break
				}
			}

			if !found {
				// 成りの選択がある場合
				if !move.IsDrop && canChoosePromote(board, move) {
					fmt.Print("成りますか？ (y/n): ")
					scanner.Scan()
					if scanner.Text() == "y" {
						move.Promote = true
					}

					// 再度チェック
					for _, lm := range legalMoves {
						if movesEqual(move, &lm) {
							move = &lm
							found = true
							break
						}
					}
				}

				if !found {
					fmt.Println("その手は指せません")
					continue
				}
			}
		}

		if move != nil {
			board.MakeMove(*move)
		}
	}
}

// 入力パース（数字のみ版）
func parseInput(input string, board *Board) *Move {
	input = strings.TrimSpace(strings.ToLower(input))

	// 持ち駒を打つ場合（例: p53, s42）
	if len(input) == 3 && !isDigit(input[0]) {
		pieces := map[byte]PieceType{
			'p': Pawn,
			's': Silver,
			'g': Gold,
			'b': Bishop,
			'r': Rook,
		}

		if pType, ok := pieces[input[0]]; ok {
			col := int(input[1]-'0') - 1 // 1→0, 2→1, ..., 5→4
			row := int(input[2]-'0') - 1 // 1→0, 2→1, ..., 5→4
			if col >= 0 && col < 5 && row >= 0 && row < 5 {
				return &Move{-1, -1, row, col, true, pType, false}
			}
		}
	}

	// 通常の移動（例: 1551）
	if len(input) == 4 && isDigit(input[0]) {
		fromCol := int(input[0]-'0') - 1 // 1→0, 2→1, ..., 5→4
		fromRow := int(input[1]-'0') - 1 // 1→0, 2→1, ..., 5→4
		toCol := int(input[2]-'0') - 1   // 1→0, 2→1, ..., 5→4
		toRow := int(input[3]-'0') - 1   // 1→0, 2→1, ..., 5→4

		if fromCol >= 0 && fromCol < 5 && fromRow >= 0 && fromRow < 5 &&
			toCol >= 0 && toCol < 5 && toRow >= 0 && toRow < 5 {
			return &Move{fromRow, fromCol, toRow, toCol, false, Empty, false}
		}
	}

	return nil
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func parseRow(s string) int {
	rows := map[string]int{
		"一": 0, "二": 1, "三": 2, "四": 3, "五": 4,
	}
	if r, ok := rows[s]; ok {
		return r
	}
	return -1
}

func getPieceName(pType PieceType) string {
	names := map[PieceType]string{
		Pawn:   "歩",
		Silver: "銀",
		Gold:   "金",
		Bishop: "角",
		Rook:   "飛",
	}
	return names[pType]
}

func movesEqual(m1, m2 *Move) bool {
	return m1.FromRow == m2.FromRow && m1.FromCol == m2.FromCol &&
		m1.ToRow == m2.ToRow && m1.ToCol == m2.ToCol &&
		m1.IsDrop == m2.IsDrop && m1.DropPiece == m2.DropPiece &&
		m1.Promote == m2.Promote
}

func canChoosePromote(board *Board, move *Move) bool {
	if move.IsDrop {
		return false
	}

	piece := board.Cells[move.FromRow][move.FromCol]
	switch piece.Type {
	case Silver, Bishop, Rook, Pawn:
		return board.canPromote(piece.Owner, move.ToRow)
	}
	return false
}
