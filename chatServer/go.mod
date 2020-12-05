module main

go 1.15

require (
	go.uber.org/zap v1.16.0
	gohipernetFake v0.0.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace gohipernetFake v0.0.0 => ../gohipernetFake
