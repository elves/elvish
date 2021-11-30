function Header(el)
  local link = pandoc.Link('', '#'..el.identifier, '',
                           {['class'] = 'anchor icon-link', ['aria-hidden'] = 'true'})
  el.content:insert(link)
  return el
end
