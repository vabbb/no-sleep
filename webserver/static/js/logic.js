function get_flow(id) {
    const checkbox = document.getElementById('check_hex')
    $.ajax({
        url: '/flow/' + id + '?hex=' + checkbox.checked,
        type: 'GET',
        success: function (response) {
            $(".list-group2").children().each(function (i) { this.remove() })
            $(".list-group2").append(response)
        },
        error: function (error) {
            console.log(error);
        }
    });
}

function update_starred() {
    $(".list-group-star").children().each(function (i) { this.remove() })
    $.ajax({
        url: '/starred',
        type: 'POST',
        success: function (response) {
            $(".list-group-star").append(response)
        },
        error: function (error) {
            console.log(error);
        }
    });
}

function change_star(icon, id) {
    var sel
    if (icon.className.match(/far/)) {
        icon.className = icon.className.replace(/far/g, 'fas')
        sel = 'true'
    } else {
        icon.className = icon.className.replace(/fas/g, 'far')
        sel = 'false'
    }
    $.ajax({
        url: '/star/' + id + '/' + sel,
        type: 'POST',
        success: function (response) {
            console.log(response);
        },
        error: function (error) {
            console.log(error);
        }
    });
    update_starred()
}

function deactivate_all() {
    $(".list-group").children().each(function (i) {
        this.className = this.className.replace(/ active/, '')
    })
    $(".list-group-star").children().each(function (i) {
        this.className = this.className.replace(/ active/, '')
    })
}

function activate(o, id) {
    deactivate_all()
    if (!o.className.match(/active/)) {
        o.className += ' active'
    }
    get_flow(id)
}

function pwn(flow_id) {
    $.ajax({
        url: '/pwn/' + flow_id,
        type: 'GET',
        success: function (response) {
            console.log(response);
        },
        error: function (error) {
            console.log(error);
        }
    });
}

const checkbox = document.getElementById('check_hex')

checkbox.addEventListener('change', (event) => {
  if (event.target.checked) {
    $('[id=hex]').each(function (i) {this.className = this.className.replace(/ d-none/, '')})
    $('[id=data]').each(function (i) {this.className += ' d-none'})
  } else {
    $('[id=data]').each(function (i) {this.className = this.className.replace(/ d-none/, '')})
    $('[id=hex]').each(function (i) {this.className += ' d-none'})
  }
})