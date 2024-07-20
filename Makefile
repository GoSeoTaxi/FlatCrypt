# Имя исполняемого файла
BINARY_NAME = flatcrypt

# Платформы для сборки
PLATFORMS = linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64

# Каталог для выходных файлов
OUTPUT_DIR = bin

# Текущая версия (вы можете изменить это на динамическое определение версии)
VERSION = 1.0.0

# Компиляционные флаги
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

all: $(PLATFORMS)

$(PLATFORMS):
	@GOOS=$(word 1, $(subst /, ,$@)) \
	GOARCH=$(word 2, $(subst /, ,$@)) \
	go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-$(word 1, $(subst /, ,$@))-$(word 2, $(subst /, ,$@)) .

clean:
	rm -rf $(OUTPUT_DIR)/*

.PHONY: all clean $(PLATFORMS)