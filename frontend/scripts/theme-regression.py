#!/usr/bin/env python3
"""Capture theme states using a disposable database and local Chromium.

Usage: python3 theme-regression.py /absolute/path/to/filebrowser /tmp/screenshots
Requires Chromium, FFmpeg and Python websocket-client; never uses an existing database.
"""
import base64
import json
import pathlib
import socket
import sys
import os
import struct
import zlib
import subprocess
import tempfile
import time
import urllib.request
import wave
import zipfile
import websocket


def port():
    with socket.socket() as sock:
        sock.bind(('127.0.0.1', 0))
        return sock.getsockname()[1]


def stop(process):
    if process and process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=8)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()


if len(sys.argv) != 3:
    raise SystemExit(__doc__)
binary = str(pathlib.Path(sys.argv[1]).resolve())
out = pathlib.Path(sys.argv[2])
out.mkdir(exist_ok=True)
password = 'LocalFontSmoke2026!'
with tempfile.TemporaryDirectory(prefix='neko-theme-regression-') as folder:
    folder = pathlib.Path(folder)
    root = folder/'files'
    root.mkdir()
    (root/'Folder').mkdir()
    (root/'notes.md').write_text('# 魔法书库\n\nLocal files and fonts.\n')
    (root/'example.json').write_text('{"magic": true}\n')
    def png_chunk(kind, data):
        return struct.pack('!I',len(data))+kind+data+struct.pack('!I',zlib.crc32(kind+data))
    (root/'pixel.png').write_bytes(b'\x89PNG\r\n\x1a\n'+png_chunk(b'IHDR',struct.pack('!2I5B',320,240,8,2,0,0,0))+png_chunk(b'IDAT',zlib.compress((b'\x00'+bytes([220,246,255])*320)*240))+png_chunk(b'IEND',b''))
    (root/'table.csv').write_text('Name,Count\nMagic,12\nBooks,24\n')
    with wave.open(str(root/'sound.wav'), 'wb') as audio:
        audio.setparams((1, 2, 8000, 0, 'NONE', 'not compressed'))
        audio.writeframes(b'\0\0' * 8000)
    subprocess.run(['ffmpeg', '-hide_banner', '-loglevel', 'error', '-f', 'lavfi',
                    '-i', 'color=c=lightblue:s=320x240:r=1', '-t', '1',
                    '-c:v', 'libvpx-vp9', '-threads', '1', '-an', str(root/'clip.webm')], check=True)
    with zipfile.ZipFile(root/'book.epub', 'w') as book:
        book.writestr('mimetype', 'application/epub+zip')
        book.writestr('META-INF/container.xml', '''<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles><rootfile full-path="content.opf" media-type="application/oebps-package+xml"/></rootfiles></container>''')
        book.writestr('content.opf', '''<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="id">
<metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:identifier id="id">theme-test</dc:identifier>
<dc:title>Library</dc:title><dc:language>en</dc:language></metadata><manifest>
<item id="chapter" href="chapter.xhtml" media-type="application/xhtml+xml"/>
<item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/></manifest>
<spine><itemref idref="chapter"/></spine></package>''')
        book.writestr('nav.xhtml', '''<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>Contents</title></head><body><nav epub:type="toc"><ol><li><a href="chapter.xhtml">Library</a></li></ol></nav></body></html>''')
        book.writestr('chapter.xhtml', '''<html xmlns="http://www.w3.org/1999/xhtml"><head><title>Library</title></head>
<body><h1>Library</h1><p>Books and files in a bright reading room.</p></body></html>''')
    pdf = bytearray(b'%PDF-1.4\n')
    objects = [b'<< /Type /Catalog /Pages 2 0 R >>', b'<< /Type /Pages /Kids [3 0 R] /Count 1 >>',
               b'<< /Type /Page /Parent 2 0 R /MediaBox [0 0 320 240] /Contents 4 0 R >>',
               b'<< /Length 0 >>\nstream\n\nendstream']
    offsets = [0]
    for index, value in enumerate(objects, 1):
        offsets.append(len(pdf))
        pdf.extend(f'{index} 0 obj\n'.encode() + value + b'\nendobj\n')
    xref = len(pdf)
    pdf.extend(b'xref\n0 5\n0000000000 65535 f \n')
    for offset in offsets[1:]:
        pdf.extend(f'{offset:010d} 00000 n \n'.encode())
    pdf.extend(f'trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n{xref}\n%%EOF\n'.encode())
    (root/'page.pdf').write_bytes(pdf)
    for fixture in root.iterdir():
        os.utime(fixture,(1761955200,1761955200))
    hashed = subprocess.run([binary, 'hash', password], check=True, capture_output=True, text=True).stdout.strip()
    server_port, debug_port = port(), port()
    origin = f'http://127.0.0.1:{server_port}'
    base = origin+'/library'
    server = chrome = None
    log = open(out/'server.log', 'w')
    chrome_log = open(out/'chrome.log', 'w')
    try:
        server = subprocess.Popen([binary, '-d', str(folder/'test.db'), '-r', str(root), '--username', 'admin', '--password', hashed, '--address', '127.0.0.1', '-p', str(server_port), '--baseURL', '/library'], cwd=folder, stdout=log, stderr=log)
        for _ in range(100):
            try:
                urllib.request.urlopen(base+'/', timeout=1).close()
                break
            except OSError:
                time.sleep(.1)
        else:
            raise AssertionError('server startup timed out')
        chrome = subprocess.Popen(['chromium', '--headless', '--no-sandbox', '--disable-gpu', '--disable-dev-shm-usage', '--disable-background-networking', '--no-first-run', '--remote-allow-origins=*', f'--remote-debugging-port={debug_port}', f'--user-data-dir={folder / "chrome"}', 'about:blank'], stdout=chrome_log, stderr=chrome_log)
        for _ in range(100):
            try:
                tabs = json.load(urllib.request.urlopen(f'http://127.0.0.1:{debug_port}/json', timeout=1))
                if tabs:
                    break
            except OSError:
                pass
            time.sleep(.1)
        ws = websocket.create_connection(next(t['webSocketDebuggerUrl'] for t in tabs if t['type']=='page'), timeout=20)
        seq = 0
        events = []

        def call(method, params=None):
            global seq
            seq += 1
            ws.send(json.dumps({'id':seq, 'method':method, 'params':params or {}}))
            while True:
                response = json.loads(ws.recv())
                if response.get('id') == seq:
                    assert 'error' not in response, response
                    return response.get('result', {})
                events.append(response)

        def ev(expression):
            result = call('Runtime.evaluate', {'expression':expression, 'awaitPromise':True, 'returnByValue':True})
            assert 'exceptionDetails' not in result, result
            return result.get('result', {}).get('value')

        def wait(expression):
            for _ in range(60):
                if ev(expression):
                    return
                time.sleep(.1)
            raise AssertionError(expression)

        def nav(path, selector):
            call('Page.navigate', {'url':base+path})
            wait('location.pathname==='+json.dumps('/library'+path)+' && document.readyState==="complete" && !!document.querySelector('+json.dumps(selector)+')')
            ev('document.fonts.ready.then(()=>true)')
            wait("!document.querySelector('#loading')")

        def shot(name):
            # Neutralize pointer hover and nondeterministic QR payloads, keeping
            # their dimensions and theme styling in the comparison.
            call('Input.dispatchMouseEvent', {'type':'mouseMoved','x':0,'y':0})
            ev("document.querySelectorAll('canvas').forEach(e=>{let c=e.getContext('2d');if(c)c.clearRect(0,0,e.width,e.height)})")
            ev("(()=>{document.querySelectorAll('.modified').forEach(e=>e.textContent='2026-11-01');document.querySelectorAll('.quota-percent').forEach(e=>e.textContent='25%');document.querySelectorAll('.quota-info span').forEach(e=>e.textContent='已使用 1 GiB / 总容量 4 GiB');document.querySelectorAll('.quota-bar').forEach(e=>{e.style.width='25%'});document.querySelectorAll('.ace_cursor').forEach(e=>e.style.visibility='hidden')})()")
            ev('document.fonts.ready.then(()=>true)')
            time.sleep(.5)
            ev("document.getAnimations().filter(a=>Number.isFinite(a.effect?.getComputedTiming().endTime)).forEach(a=>a.finish())")
            ev("new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve)))")
            time.sleep(.1)
            styles=ev("(()=>{let props=['display','position','top','right','bottom','left','width','height','min-width','max-width','min-height','max-height','margin','padding','gap','flex','flex-direction','align-items','justify-content','grid-template-columns','color','background-color','background-image','background-size','background-position','border','border-radius','box-shadow','text-shadow','font-family','font-size','font-weight','line-height','letter-spacing','text-align','white-space','overflow','opacity','z-index','transform','outline','outline-offset','appearance','-webkit-font-smoothing','text-rendering','font-optical-sizing','font-synthesis','font-stretch','font-kerning','font-feature-settings','font-variation-settings','filter','backdrop-filter','mix-blend-mode','will-change','contain','clip-path'];return [...document.querySelectorAll('body *')].filter(e=>e.getBoundingClientRect().width&&e.getBoundingClientRect().height).map(e=>{let path=[],n=e;while(n&&n!==document.body){path.unshift(n.tagName.toLowerCase()+':'+[...n.parentElement.children].indexOf(n));n=n.parentElement}let br=e.getBoundingClientRect();let result={box:{x:br.x,y:br.y,w:br.width,h:br.height},path:path.join('/'),tag:e.tagName,cls:e.className?.baseVal??e.className};for(let pseudo of ['', '::before','::after']){let s=getComputedStyle(e,pseudo||null);if(pseudo&&['none','normal'].includes(s.content))continue;result[pseudo||'element']=Object.fromEntries(props.map(p=>[p,s.getPropertyValue(p).replaceAll(location.origin,'ORIGIN')]))}return result})})()")
            (out/(name+'.json')).write_text(json.dumps(styles,ensure_ascii=False,indent=2))
            image = call('Page.captureScreenshot', {'format':'png'})
            (out/(name+'.png')).write_bytes(base64.b64decode(image['data']))

        def fonts(selector):
            node = call('DOM.querySelector', {'nodeId':call('DOM.getDocument')['root']['nodeId'], 'selector':selector})['nodeId']
            return call('CSS.getPlatformFontsForNode', {'nodeId':node})['fonts']

        call('Page.enable')
        call('Network.enable')
        call('DOM.enable')
        call('CSS.enable')
        call('Network.setCacheDisabled', {'cacheDisabled':True})
        call('Network.setBlockedURLs', {'urls':['https://*', '*fonts.googleapis.com*', '*fonts.gstatic.com*', '*cdn.jsdelivr.net*']})
        call('Emulation.setDeviceMetricsOverride', {'width':1440, 'height':1000, 'deviceScaleFactor':1, 'mobile':False})
        def style(selector):
            return ev('(()=>{let e=document.querySelector('+json.dumps(selector)+');if(!e)return null;let s=getComputedStyle(e),r=e.getBoundingClientRect();return{w:r.width,h:r.height,x:r.x,y:r.y,color:s.color,bg:s.backgroundColor,border:s.borderColor,borderWidth:s.borderWidth,shadow:s.boxShadow,animation:s.animationName,transform:s.transform,appearance:s.appearance,outline:s.outline,checked:e.checked,disabled:e.disabled}})()')

        def click(selector):
            assert ev('!!document.querySelector('+json.dumps(selector)+')'), selector
            ev('document.querySelector('+json.dumps(selector)+').click()')
            time.sleep(.18)

        def action(icon, scope='header'):
            ev('(()=>{let b=[...document.querySelectorAll('+json.dumps(scope+' button, #file-selection button')+')].find(e=>e.querySelector("i")?.textContent==='+json.dumps(icon)+');if(!b)throw new Error("action missing");b.click()})()')
            time.sleep(.2)

        def tap(selector):
            pos=ev('(()=>{let e=document.querySelector('+json.dumps(selector)+'),r=e.getBoundingClientRect(),x=r.x+r.width/2,y=r.y+r.height/2;return {x,y,hit:e===document.elementFromPoint(x,y)||e.contains(document.elementFromPoint(x,y))}})()')
            assert pos.pop('hit'), ('covered control',selector,pos)
            call('Input.dispatchMouseEvent',{'type':'mousePressed','button':'left','clickCount':1,**pos})
            call('Input.dispatchMouseEvent',{'type':'mouseReleased','button':'left','clickCount':1,**pos})
            time.sleep(.2)

        def close_dialog():
            assert ev("!!document.querySelector('#modal-background .prompt-icon')")
            assert ev("getComputedStyle(document.querySelector('#modal-background')).backgroundImage.includes('radial-gradient')")
            tap('#modal-background .prompt-close')
            wait("!document.querySelector('#modal-background')")

        nav('/login', '#login img')
        time.sleep(.4)
        first=style('#login img')
        time.sleep(.7)
        second=style('#login img')
        assert first['animation']=='none' and first['transform']=='none' and first['w']>=155, first
        assert first['x']==second['x'] and first['y']==second['y'], (first,second)
        print('STATIC LARGE LOGIN LOGO',first,flush=True)
        shot('login-desktop')
        for height in [844,667]:
            call('Emulation.setDeviceMetricsOverride',{'width':390,'height':height,'deviceScaleFactor':1,'mobile':True})
            assert ev("(()=>{let f=document.querySelector('#login form'),r=f.getBoundingClientRect();return r.top>=0 && document.documentElement.scrollWidth<=innerWidth})()")
            shot('login-mobile-'+str(height))
        call('Emulation.setDeviceMetricsOverride',{'width':1440,'height':1000,'deviceScaleFactor':1,'mobile':False})
        ev('''(()=>{let u=document.querySelector('#login input[type=text]'),p=document.querySelector('#login input[type=password]');u.value='admin';u.dispatchEvent(new Event('input',{bubbles:true}));p.value='''+json.dumps(password)+''';p.dispatchEvent(new Event('input',{bubbles:true}));document.querySelector('#login input[type=submit]').click()})()''')
        wait('!!document.querySelector("#listing")')
        nav('/settings/profile','select[name=selectLanguage]')
        ev("(()=>{let s=document.querySelector('select[name=selectLanguage]');s.value='zh-cn';s.dispatchEvent(new Event('change',{bubbles:true}));document.querySelector('input[name=submitProfile]').click()})()")
        time.sleep(.4)
        nav('/files/','#listing')
        assert ev("document.querySelector('.header-logo-img').parentElement.classList.contains('header-brand-wrap')")
        assert not ev("!!document.querySelector('.header-logo-badge,.header-logo-star')")
        logo=style('.header-logo-img')
        assert logo['w']>=50 and logo['borderWidth']=='0px' and logo['shadow']=='none',logo
        click('.item[aria-label="notes.md"]')
        toolbar=ev("[...document.querySelectorAll('header .action')].filter(e=>e.offsetParent).map(e=>({icon:e.dataset.icon,tone:e.dataset.tone,color:getComputedStyle(e.querySelector('i')).color,bg:getComputedStyle(e).backgroundColor}))")
        colors={e['color'] for e in toolbar}
        assert len(colors)>=4 and 'rgb(30, 42, 74)' not in colors and 'rgb(0, 0, 0)' not in colors, toolbar
        sidebar=ev("[...document.querySelectorAll('#sidebar .action:not(.active) i')].map(e=>({icon:e.textContent,color:getComputedStyle(e).color}))")
        assert len({e['color'] for e in sidebar})>=4,sidebar
        print('DIRECT HEADER LOGO',logo,flush=True)
        print('TOOLBAR PALETTE',toolbar,flush=True)
        print('SIDEBAR PALETTE',sidebar,flush=True)
        shot('file-toolbar-sidebar')
        for icon in ['share','content_copy','forward','delete','file_download']:
            selector='header .action[data-icon="'+icon+'"]'
            pos=ev('(()=>{let r=document.querySelector('+json.dumps(selector)+').getBoundingClientRect();return{x:r.x+r.width/2,y:r.y+r.height/2}})()')
            call('Input.dispatchMouseEvent',{'type':'mouseMoved',**pos})
            time.sleep(.15)
            print('HOVER',icon,style(selector),flush=True)

        nav('/settings/profile','input[name=hideDotfiles]')
        checkbox='input[name=hideDotfiles]'
        initial=style(checkbox)
        assert initial['appearance']=='none' and initial['w']==20 and initial['borderWidth']=='2px',initial
        ev("document.querySelector('input[name=hideDotfiles]').focus()")
        for event in ['keyDown','keyUp']:
            call('Input.dispatchKeyEvent',{'type':event,'key':' ','code':'Space','windowsVirtualKeyCode':32})
        assert style(checkbox)['checked']!=initial['checked']
        assert '3px' in style(checkbox)['outline']
        print('KEYBOARD CHECKBOX',style(checkbox),flush=True)
        shot('profile-checkbox-focus')
        click(checkbox)
        assert style(checkbox)['checked']==initial['checked']
        ev("document.querySelector('input[name=hideDotfiles]').indeterminate=true")
        assert 'M5%2010h10' in ev("getComputedStyle(document.querySelector('input[name=hideDotfiles]')).backgroundImage") or 'M5 10h10' in ev("getComputedStyle(document.querySelector('input[name=hideDotfiles]')).backgroundImage")
        ev("document.querySelector('input[name=hideDotfiles]').indeterminate=false")
        call('Emulation.setEmulatedMedia',{'features':[{'name':'forced-colors','value':'active'}]})
        assert style(checkbox)['appearance']=='auto'
        call('Emulation.setEmulatedMedia',{'features':[]})
        nav('/settings/global','#minimumPasswordLength')
        disabled='.dashboard input[type=checkbox]:disabled'
        before=style(disabled)
        assert before and before['disabled']
        click(disabled)
        assert style(disabled)['checked']==before['checked']
        print('DISABLED CHECKBOX',before,flush=True)
        shot('global-checkboxes')

        for width in [1440,390]:
            call('Emulation.setDeviceMetricsOverride',{'width':width,'height':1000 if width>500 else 844,'deviceScaleFactor':1,'mobile':width<500})
            nav('/files/','#listing')
            click('.item[aria-label="notes.md"]')
            for icon,prompt,tone in [('file_upload','upload','pink'),('share','share','cyan'),('content_copy','copy','purple'),('forward','move','cyan'),('info','info','cobalt'),('delete','delete','pink')]:
                action(icon)
                wait("!!document.querySelector('#modal-background .card')")
                assert ev("document.querySelector('#modal-background').dataset.tone")==tone,(prompt,width)
                assert ev("document.documentElement.scrollWidth<=innerWidth+1"),(prompt,width)
                assert ev("(()=>{let e=document.querySelector('#modal-background .card');return e.scrollWidth<=e.clientWidth+1 && e.getBoundingClientRect().bottom<=innerHeight})()"),(prompt,width)
                print('DIALOG',width,prompt,style('#modal-background .card-title'),flush=True)
                shot(prompt+'-'+str(width))
                if prompt=='share':
                    ev("[...document.querySelectorAll('#modal-background button')].find(e=>e.querySelector('i')?.textContent==='add')?.click()")
                    time.sleep(.2)
                    shot('share-settings-'+str(width))
                close_dialog()
            if width<500:
                action('menu')
                wait("!!document.querySelector('#sidebar.active')")
                shot('mobile-sidebar')
                # The sidebar overlay closes the same drawer state.
                ev("document.querySelector('nav#sidebar').previousElementSibling.click()")
                time.sleep(.2)
            nav('/settings/profile','input[name=hideDotfiles]')
            assert ev('document.documentElement.scrollWidth<=innerWidth+1')
            shot('profile-'+str(width))
        call('Emulation.setDeviceMetricsOverride',{'width':1440,'height':1000,'deviceScaleFactor':1,'mobile':False})
        nav('/files/pixel.png','#previewer .preview')
        wait("!!document.querySelector('#previewer header .action[data-tone=pink] i')")
        assert style('#previewer header .action[data-tone=pink] i')['color']=='rgb(201, 59, 107)'
        assert style('#previewer header .action[data-tone=cyan] i')['color']=='rgb(20, 107, 150)'
        shot('image-preview-toolbar')
        nav('/files/notes.md','#editor')
        shot('text-editor-toolbar')
        for width,height in [(1440,1000),(390,844),(320,667),(736,414)]:
            call('Emulation.setDeviceMetricsOverride',{'width':width,'height':height,'deviceScaleFactor':1,'mobile':width<=736})
            nav('/files/','#listing')
            action('check_circle')
            wait("!!document.querySelector('.selection-dock #multiple-selection')")
            click('.item[aria-label="notes.md"]')
            click('.item[aria-label="example.json"]')
            time.sleep(.3)
            assert ev('document.documentElement.scrollWidth<=innerWidth+1'),('overflow',width)
            assert ev("(()=>{let d=document.querySelector('.selection-dock').getBoundingClientRect();return d.left>=0 && d.right<=innerWidth && d.bottom<=innerHeight})()"),('dock',width)
            assert ev("parseFloat(getComputedStyle(document.querySelector('.file-listing')).paddingBottom)>=document.querySelector('.selection-dock').getBoundingClientRect().height"),('reserved space',width)
            if width<=736:
                assert style('#file-selection')['bg']=='rgb(255, 255, 255)'
                assert ev("(()=>{let a=document.querySelector('.selection-summary').getBoundingClientRect(),b=document.querySelector('.selection-actions').getBoundingClientRect();return a.bottom<=b.top})()"),('summary overlaps actions',width)
                # Scroll cards behind the dock, then test the actual pointer hit target.
                ev('window.scrollTo(0,document.documentElement.scrollHeight/2)')
                time.sleep(.2)
                tap('#file-selection .action[data-icon=content_copy]')
                wait("!!document.querySelector('#modal-background')")
                close_dialog()
            shot('multiple-dock-'+str(width))
            tap('.selection-dismiss')
            wait("!document.querySelector('#multiple-selection')")
            if width<=736:
                action('more_vert')
                wait("!!document.querySelector('#dropdown.active')")
                assert ev("getComputedStyle(document.querySelector('header .overlay')).backgroundImage.includes('radial-gradient')")
                assert style('#dropdown')['bg']=='rgb(255, 255, 255)'
                assert ev("(()=>{let r=document.querySelector('#dropdown').getBoundingClientRect();return r.left>=0 && r.right<=innerWidth && r.bottom<=innerHeight})()"),('menu',width)
                shot('menu-backdrop-'+str(width))
                # Open a prompt through the menu, then close with its new header button.
                tap('#dropdown .action[data-icon=info]')
                wait("!!document.querySelector('#modal-background')")
                close_dialog()
                # Closing the prompt reveals the still-open menu, whose backdrop dismisses it.
                if ev("!!document.querySelector('#dropdown.active')"):
                    pos={'x':width-8,'y':height-8}
                    assert ev('document.elementFromPoint('+str(width-8)+','+str(height-8)+').matches("header .overlay")')
                    call('Input.dispatchMouseEvent',{'type':'mousePressed','button':'left','clickCount':1,**pos})
                    call('Input.dispatchMouseEvent',{'type':'mouseReleased','button':'left','clickCount':1,**pos})
                    wait("!document.querySelector('#dropdown.active')")
        print('PASS: prompt icons, close buttons and halftone backdrops; responsive selection docks, hit targets and mobile menus at 1440/390/320/736px.',flush=True)
        share = ev("fetch('/library/api/share/',{method:'POST',headers:{'X-Auth':localStorage.getItem('jwt')},body:'{}'}).then(async r=>{if(!r.ok)throw new Error('temporary share failed');return r.json()})")
        assert share.get('hash'),share
        for width in [1440,390]:
            call('Emulation.setDeviceMetricsOverride',{'width':width,'height':1000 if width>736 else 844,'deviceScaleFactor':1,'mobile':width<=736})
            nav('/share/'+share['hash'],'#shareList')
            action('check_circle')
            wait("!!document.querySelector('.selection-dock #multiple-selection')")
            assert ev('document.documentElement.scrollWidth<=innerWidth+1')
            assert ev("parseFloat(getComputedStyle(document.querySelector('.file-listing')).paddingBottom)>=document.querySelector('.selection-dock').getBoundingClientRect().height")
            shot('public-share-notice-'+str(width))
            tap('.selection-dismiss')
            wait("!document.querySelector('.selection-dock')")
        print('PASS: shared notice and reserved scroll space on the public share page.',flush=True)
        for width in [1440,390]:
            call('Emulation.setDeviceMetricsOverride',{'width':width,'height':1000 if width>736 else 844,'deviceScaleFactor':1,'mobile':width<=736})
            nav('/files/','#listing')
            for _ in range(3):
                mode=ev("document.querySelector('#listing').className")
                shot('view-'+mode.replace(' ','-')+'-'+str(width))
                ev("[...document.querySelectorAll('header button')].find(e=>['view_list','grid_view','view_module','view_comfy','collections','photo_library','apps'].includes(e.dataset.icon))?.click()")
                time.sleep(.3)
            action('search',scope='header')
            wait("!!document.querySelector('#search.active')")
            shot('search-'+str(width))
            ev("document.querySelector('#search input').value='notes';document.querySelector('#search input').dispatchEvent(new Event('input',{bubbles:true}))")
            time.sleep(.1)
            ev("document.querySelector('#search input').dispatchEvent(new KeyboardEvent('keyup',{key:'Enter',bubbles:true}))")
            wait("!!document.querySelector('#search #result li')")
            time.sleep(.8)
            shot('search-results-'+str(width))
            nav('/files/table.csv','#previewer')
            wait("!!document.querySelector('.csv-viewer')")
            shot('csv-preview-'+str(width))
            nav('/files/notes.md','#editor.ace_editor')
            shot('editor-'+str(width))
            action('preview')
            time.sleep(.3)
            shot('markdown-preview-'+str(width))
            for filename, selector, label in [
                ('sound.wav', '#previewer audio', 'audio'),
                ('clip.webm', '#previewer .video-js', 'video'),
                ('page.pdf', '#previewer .pdf', 'pdf'),
                ('book.epub', '#previewer .epub-reader iframe', 'epub'),
            ]:
                nav('/files/'+filename, selector)
                if label in ('audio', 'video'):
                    ev("document.querySelectorAll('audio,video').forEach(e=>e.pause())")
                time.sleep(.5)
                if label == 'video':
                    assert ev("(()=>{let e=document.querySelector('.video-js'),p=e.parentElement;return Math.abs(e.clientWidth-p.clientWidth)<=1&&Math.abs(e.clientHeight-p.clientHeight)<=1})()"), 'video must fill its preview surface'
                if label == 'epub':
                    wait("!!document.querySelector('.epub-reader .tocAreaButton.active')")
                    assert ev("getComputedStyle(document.querySelector('.epub-reader .tocAreaButton.active')).color") == 'rgb(78, 141, 245)'
                shot(label+'-preview-'+str(width))
        # Check against the content box, not just the viewport: legacy table
        # gutters can overflow a card while remaining within the page bounds.
        long_share_name = 'Library_' + 'long_path_' * 16
        (root/long_share_name).mkdir()
        ev('fetch('+json.dumps('/library/api/share/'+long_share_name+'/')+',{method:"POST",headers:{"X-Auth":localStorage.getItem("jwt")},body:"{}"}).then(r=>{if(!r.ok)throw new Error("temporary share failed");return r.json()})')
        for width in [1440,1024,800,390,320]:
            call('Emulation.setDeviceMetricsOverride',{'width':width,'height':1000 if width>736 else 844,'deviceScaleFactor':1,'mobile':width<=736})
            nav('/settings/shares','.shares-table')
            assert ev("(()=>{let c=document.querySelector('.share-manager'),p=c.parentElement,s=getComputedStyle(p);return Math.abs(c.getBoundingClientRect().width-(p.clientWidth-parseFloat(s.paddingLeft)-parseFloat(s.paddingRight)))<=1})()"), 'share management must use the available column width'
            assert ev("(()=>{let e=document.querySelector('.share-list-content'),t=e.querySelector('table'),r=e.getBoundingClientRect(),q=t.getBoundingClientRect(),s=getComputedStyle(e);return q.left>=r.left+parseFloat(s.paddingLeft)-1&&q.right<=r.right-parseFloat(s.paddingRight)+1&&e.scrollWidth<=e.clientWidth+1&&e.scrollHeight<=e.clientHeight+1})()"), 'share list must remain inside its padded content box without nested scrolling'
            assert ev("getComputedStyle(document.querySelector('.share-manager')).isolation==='isolate'&&getComputedStyle(document.querySelector('.share-manager'),'::before').zIndex==='-1'"), 'decorative frame must stay behind card contents'
            assert ev('document.documentElement.scrollWidth<=innerWidth+1')
            shot('share-management-'+str(width))
        errors=[e['params']['response'] for e in events if e.get('method')=='Network.responseReceived' and e['params']['response']['status']>=400]
        assert not errors,errors
        print('PASS: theme states captured; responsive controls, dialogs, local assets and keyboard interactions verified under /library.',flush=True)
        ws.close()
    finally:
        stop(chrome)
        stop(server)
        chrome_log.close()
        log.close()
