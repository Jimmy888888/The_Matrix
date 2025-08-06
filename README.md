# The_Matrix
Use vibe coding to build the digital rain in The Matrix, the picture below is my CLI view

<img width="1439" height="863" alt="截圖 2025-08-06 晚上9 49 05" src="https://github.com/user-attachments/assets/2cd1267f-c0c8-4da7-9f92-5ec47d4d114c" />


reference: https://github.com/hylarucoder/codebench/blob/master/T002-matrix/PROMPT-golang.md?plain=1

## Learn:
go channel

	time Ticker
 
	os Stdin / Signal
	

go routine

## Issues:

如何追蹤新生出來的columns?

draw()裡面的 清除機制沒有跟著speed調整？

reset()如何做到清除畫面？ 

reset只重置columns的位置跟內容 不清除畫面

ANSI 轉義碼

解釋一個column的runtime

Columns 維護一個column slice , 在每次time channel trigger時打印並更新
