# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a sophisticated cryptocurrency quantitative trading system (deeptrade) that specializes in automated ETH futures trading on Binance. The system uses AI-powered analysis with multiple LLM providers (DeepSeek, Qwen) for decision making and implements advanced technical indicators with real-time market data processing.

## Key Architecture Components

### Core System Structure
- **main.go**: Entry point with trading timer system (operates 15:00-02:59)
- **task/**: Core trading logic and analysis modules
- **binance/**: Binance API integration with futures client
- **indicators/**: Technical analysis indicators (ATR, Bollinger Bands, MACD, RSI, SMA/EMA)
- **utils/**: Utility functions including LLM integration and email notifications
- **conf/**: Configuration management with TOML-based settings

### Trading Process Flow
1. **Market Data Collection** (`task/market_data.go`): Concurrently fetches ticker, klines, orderbook, positions, account info
2. **Technical Analysis** (`task/analysis.go`): Prepares data and calls LLM for analysis
3. **Signal Generation** (`task/execution.go`): Converts LLM output to trading signals
4. **Trade Execution** (`binance/futures_client.go`): Executes trades on Binance

### AI Integration
- Uses `cloudwego/eino` framework for LLM integration
- Supports multiple providers: DeepSeek, Qwen, OpenAI-compatible models
- Configuration via `conf/config.toml` `[llm]` section

## Common Development Commands

### Building and Running
```bash
# Build main application
go build main.go

# Run directly
./main

# Run with background process (production style)
./run.sh
```

### Testing
```bash
# Run all tests
go test -v ./task

# Run specific test
go test -v ./task -run TestGetMarketData

# Run with timeout for market data tests
go test -v ./task -run TestGetMarketData -timeout 60s
```

### Development Workflow
```bash
# Clean build
rm ./main && go build main.go

# Kill existing process and rebuild
pkill -f './main' && go build main.go
```

## Configuration Management

### Environment Configuration
- **Testnet**: Set `current_environment = "testnet"` in config.toml
- **Production**: Set `current_environment = "production"` in config.toml
- System automatically switches environment based on this setting

### Key Configuration Files
- `conf/config.toml`: Main configuration including API keys, trading parameters
- `conf/config.go`: Configuration structure and loading logic

### Important Config Sections
- `[binance]`: API keys and endpoints for testnet/production
- `[llm]`: Multiple LLM provider configurations
- `[trading]`: Trading parameters like `trigger_time` (minutes between cycles)

## Data Structures and Types

### Core Market Data (`task/types.go`)
- `MarketData`: Comprehensive market data container
- `TradingSignal`: AI-generated trading signals with confidence scores
- `TechnicalData`: Processed technical indicators

### API Integration (`binance/types.go`)
- Complete Binance futures API type definitions
- Standardized market data structures

## Development Guidelines

### Testing Market Data
The system includes comprehensive market data tests. Use the provided test patterns:
- Tests use real API endpoints (configured for testnet)
- Market data tests require network connectivity
- Use timeout flags for network-dependent tests

### LLM Integration
- LLM calls are abstracted through `utils.Run()`
- Response parsing is centralized in `task/analysis.go`
- Multiple model providers supported with fallbacks

### Error Handling
- Network calls include retry logic with configurable max retries
- System logs all trading decisions and errors
- Email notifications for critical errors (configured in utils)

### Technical Indicators
- Located in `indicators/` directory
- Each indicator has its own file with clear interfaces
- Used for both analysis and LLM context generation

## Safety and Risk Management

### Trading Hours
- System only trades between 15:00-02:59 (configurable in `IsWork()`)
- Outside these hours, the system sleeps but monitors positions
- Continues running if positions are open, even outside trading hours

### Position Management
- Position sizes calculated as percentage of available margin (`PositionPercent`)
- Cross-margin mode with configurable leverage
- Automatic position size calculations based on account equity

### API Security
- Separate configurations for testnet and production environments
- API keys stored in configuration file (not hardcoded)
- Supports proxy configuration for network access

## Dependencies and Ecosystem

### Key External Dependencies
- `github.com/8treenet/freedom`: Application framework
- `github.com/cloudwego/eino`: LLM integration framework
- `github.com/gorilla/websocket`: WebSocket support for real-time data

### Go Version
- Requires Go 1.24.0+ (specified in go.mod)
- Uses toolchain go1.24.7

## Debugging and Monitoring

### Logging
- Comprehensive logging throughout the system with structured messages
- Chinese log messages for local development context
- Error logging includes stack traces where applicable

### Market Data Debugging
- Use `TestGetMarketData` to verify API connectivity
- Check log outputs for real-time market data retrieval
- System prints current price, account balance, and position status

### Configuration Validation
- System validates configuration on startup
- Missing or invalid configurations are logged
- Environment-specific settings are enforced