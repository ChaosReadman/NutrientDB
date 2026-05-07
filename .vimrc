" NutrientDB プロジェクト固有のVim設定

" Goファイル用: Goの標準（gofmt）に従い、タブを使用し、保存時にフォーマットする設定を推奨
autocmd FileType go setlocal noexpandtab tabstop=4 shiftwidth=4
" vim-goプラグインを入れている場合、保存時に自動gofmtがかかるようにすると便利です
let g:go_fmt_autosave = 1

" Pythonファイル用: PEP 8に従い、スペース4つを使用
autocmd FileType python setlocal expandtab tabstop=4 shiftwidth=4

" HTMLテンプレート用: ネストを深くしすぎないよう、スペース2つが視認性良くなります
autocmd FileType html setlocal expandtab tabstop=2 shiftwidth=2

" 検索のハイライトをサッと消すためのマッピング（vi使いには必須級です）
nnoremap <silent> <Esc><Esc> :<C-u>set nohlsearch<Return>

" ファイルブラウザ（netrw）を使いやすくする設定
let g:netrw_banner = 0
let g:netrw_liststyle = 3