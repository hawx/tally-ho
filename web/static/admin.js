const notification = document.getElementById('notification');

notification.querySelector('.delete').onclick = function() {
  notification.classList.add('is-hidden');
}

function notify(href) {
  const a = notification.querySelector('a');
  a.textContent = href;
  a.href = href;

  notification.classList.remove('is-hidden');
}

function postJSON(data) {
  return fetch(micropub, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${ token }`,
    },
    body: JSON.stringify(data),
  })
    .then(resp => {
      if (resp.status === 201) {
        const location = resp.headers.get('Location');
        notify(location);
      } else {
        console.log('There was a problem', resp);
      }
    });
}

function setupTabs(tabs, tabbed) {
  window.onhashchange = function () {
    const hash = window.location.hash;

    for (const tab of tabs) {
      tab.classList.remove('is-active');

      if (tab.firstChild.hash === hash) {
        tab.classList.add('is-active');
      }
    }

    for (const container of tabbed) {
      container.classList.add('is-hidden');

      if (container.id === hash.slice(1)) {
        container.classList.remove('is-hidden');
      }
    }
  }

  window.onhashchange();

  if (!window.location.hash) {
    window.location.hash = tabs[0].firstChild.hash;
  }
}

function askForFile() {
  return new Promise((resolve, reject) => {
    const modal = document.getElementById('modal');
    const fileUpload = modal.querySelector('input[type=file]');
    const doUpload = modal.querySelector('.button.is-success');

    modal.classList.add('is-active');

    fileUpload.onchange = function() {
      if (fileUpload.files.length === 1) {
        document.getElementById('file-name').textContent = fileUpload.files[0].name;
      }
    };

    doUpload.onclick = function() {
      const formData = new FormData();

      formData.append('file', fileUpload.files[0]);

      fetch(media, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${ token }`,
        },
        body: formData,
      })
        .then(resp => {
          resolve(resp.headers.get('Location'));
          modal.classList.remove('is-active');
        })
        .catch(err => reject(err));
    };
  });
}

function setupPostForm(el) {
  const form = el.querySelector('form');
  const content = form.elements['content'];
  const editor = el.querySelector('.editor');

  pell.init({
    element: editor,
    defaultParagraphSeparator: 'p',
    onChange(data) {
      content.value = data;
    },
    actions: [
      'bold',
      'italic',
      'underline',
      'strikethrough',
      'heading1',
      'heading2',
      'paragraph',
      'quote',
      'olist',
      'ulist',
      'code',
      'line',
      'link',
      {
        name: 'image',
        icon: '&#128247;',
        title: 'Image',
        result: () => askForFile()
          .then(location => pell.exec('insertImage', location))
          .catch(err => console.warn(err)),
      },
      {
        name: 'clear',
        icon: '-',
        title: 'Clear formatting',
        result: () => pell.exec('removeFormat'),
      },
    ],
  });

  const editorForm = document.getElementById('editorform');

  editorform.addEventListener('submit', function(e) {
    e.preventDefault();

    postJSON({
      type: ['h-entry'],
      properties: {
        name: [editorForm.elements['name'].value],
        content: [
          { html: editorForm.elements['content'].value },
        ],
      },
    });
  });
}

function setupJSONForm(el) {
  const form = el.querySelector('form');

  form.addEventListener('submit', function(e) {
    e.preventDefault();

    let data;
    try {
      data = JSON.parse(form.elements['content'].value);
    } catch (err) {
      window.alert(err);
      return;
    }

    postJSON(data);
  });
}

setupPostForm(document.getElementById('post'));
setupJSONForm(document.getElementById('json'));
setupTabs(document.querySelectorAll('.tabs li'), document.querySelectorAll('.tabbed'));
