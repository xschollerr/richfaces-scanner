# RichFaces Scanner

Uma ferramenta em Go para detectar aplicações web que utilizam RichFaces/Seam Framework.

## Características

- Suporte a múltiplas URLs de entrada
- Verifica múltiplas portas e caminhos comuns
- Processamento paralelo para maior velocidade
- Detecção de diversos padrões do RichFaces e Seam
- Ignora certificados SSL inválidos
- Salva resultados em arquivo

## Uso

```bash
./1rich -i lista.txt [-o resultados.txt] [-w 20]
```

Onde:
- `-i lista.txt`: Arquivo com lista de URLs/IPs para verificar
- `-o resultados.txt`: Arquivo de saída (opcional, padrão: resultados.txt)
- `-w 20`: Número de workers paralelos (opcional, padrão: 20)

## Instalação

```bash
git clone https://github.com/seu-usuario/richfaces-scanner
cd richfaces-scanner
go build 1rich.go
```

## Autor

xscholler - since 2000-2024

## Licença

Este projeto está sob a licença MIT.
