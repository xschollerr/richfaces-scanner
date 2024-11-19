# Configurar seu usuário Git (se ainda não fez)
git config --global user.name "xschollerr"
git config --global user.email "xstpl@riseup.net"

# Remover o repositório Git atual e começar de novo
rm -rf .git

# Inicializar novo repositório
git init

# Adicionar os arquivos
git add .

# Fazer o primeiro commit
git commit -m "Commit inicial: Scanner RichFaces"

# Adicionar o remote e fazer push
git remote add origin https://github.com/xschollerr/richfaces-scanner.git
git branch -M main
git push -u origin main
