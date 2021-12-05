function Header(el)
  local id = el.identifier
  if id == '' then return el end
  local link = pandoc.Link('', '#'..id, '',
                           {['class'] = 'anchor icon-link', ['aria-hidden'] = 'true'})
  el.content:insert(link)
  el.attributes['onclick'] = ''
  return el
end
