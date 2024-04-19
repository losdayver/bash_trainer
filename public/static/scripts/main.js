const palette = document.getElementById('palette');
const prompt = document.getElementById('prompt');
const queue = document.getElementById('queue');

const execute = document.getElementById('prompt-buttons-execute');
const save = document.getElementById('prompt-buttons-save');

const modal_username = document.getElementById('modal-username');
const modal_password = document.getElementById('modal-password');
const modal_login = document.getElementById('modal-login');

const modal = document.getElementById('myModal');

// Host api URL
const apiUrl = `${window.location.protocol}//${window.location.host}/bash_trainer/api`;
// Host static dir URL
const staticUrl = `${window.location.protocol}//${window.location.host}/bash_trainer/public/static`;

// List of Command objects
let commands = []

// Token from user logging in
let userToken = '';

// Represents the command with text and boolean value, which is true if command has args
class Command {
    constructor(text, hasArgs) {
        this.text = text;
        this.hasArgs = !!hasArgs;
    }
}

// Event happens on cursor dragging over prompt
function onDragStart(event) {
    event
        .dataTransfer
        .setData('text/plain', event.target.id);
}

// Event happens on cursor dragging over prompt
function onDragOver(event) {
    event.preventDefault();

    const dropzone = event.target;

    if (dropzone.classList.contains('command-prompt')) {
        dropzone.classList.add('command-prompt-draggedover');
    }
}

// Event happens on cursor leaving prompt
function onDragLeave(event) {
    event.preventDefault();

    const dropzone = event.target;

    if (dropzone.classList.contains('command-prompt')) {
        dropzone.classList.remove('command-prompt-draggedover');
    }
}

// Event happens on drag-and-dropping commands from palette to prompt
function onDrop(event) {
    const id = event
        .dataTransfer
        .getData('text');

    const paletteCommand = document.getElementById(id);
    const dropzone = event.target;

    if (dropzone === document.getElementById('prompt')) {
        const command = constructPromptCommand(paletteCommand.data);
        dropzone.appendChild(command);
    }
    else if (dropzone.classList.contains('command-prompt')) {
        const command = constructPromptCommand(paletteCommand.data);
        dropzone.classList.remove('command-prompt-draggedover');
        dropzone.after(command);
    }

    event
        .dataTransfer
        .clearData();
}

// Returns new node to be appended to palette
function constructPaletteCommand(text) {
    let command = document.createElement('div');

    command.className = 'command-palette';
    command.id = `command-palette-${text.replace(/ /g, '-')}`
    command.data = text;
    command.innerText = text;

    command.draggable = true;
    command.addEventListener('dragstart', event => onDragStart(event));

    command.addEventListener('click', () => {
        const commandPrompt = constructPromptCommand(command.data);
        prompt.appendChild(commandPrompt);
    });

    return command;
}

// Returns new node to be appended to prompt
function constructPromptCommand(text) {
    const command = document.createElement('div');

    const spanText = document.createElement('span');
    spanText.className = 'command-prompt-text';
    spanText.innerText = text;

    command.appendChild(spanText);

    if (commands.find(command => text === command.text).hasArgs) {
        const args = document.createElement('input');
        args.className = 'command-prompt-args';
        command.appendChild(args);
    }

    command.className = 'command-prompt';
    //command.innerText = text;

    command.addEventListener('click', event => {
        if (event.target === command) {
            command.remove();
        }
    });

    return command;
}

// Appends new task to queue
function appendTaskRunning(token, text) {
    const commandQueueRunning = document.createElement('div');
    commandQueueRunning.className = 'command-queue-running';
    commandQueueRunning.data = token;

    const commandQueueStatus = document.createElement('div');
    commandQueueStatus.classList.add('command-queue-status');

    const commandQueueStatusSymbol = document.createElement('img');
    commandQueueStatusSymbol.src = `${staticUrl}/media/vector/three-dots.svg`;
    commandQueueStatusSymbol.classList.add('command-queue-status-symbol');

    const commandQueuePrompt = document.createElement('div');
    commandQueuePrompt.classList.add('command-queue-prompt');
    commandQueuePrompt.innerText = text

    const commandQueueStatusText = document.createElement('p');
    commandQueueStatusText.classList.add('command-queue-status-text');
    commandQueueStatusText.innerText = 'Выполняется...';

    const commandQueueCross = document.createElement('img');
    commandQueueCross.src = `${staticUrl}/media/vector/x.svg`;
    commandQueueCross.classList.add('command-queue-cross');

    commandQueueRunning.appendChild(commandQueueStatus);
    commandQueueStatus.appendChild(commandQueueStatusSymbol);
    commandQueueStatus.appendChild(commandQueueStatusText);

    commandQueueRunning.appendChild(commandQueuePrompt);
    commandQueueRunning.appendChild(commandQueueCross);

    queue.appendChild(commandQueueRunning);

    commandQueueCross.addEventListener('click', () => {
        if (commandQueueRunning.className != 'command-queue-running') {
            commandQueueRunning.remove();
        }
    })

    let interval;
    interval = setInterval(() => {
        fetch(`${apiUrl}/task/` + token)
            .then(data => data.json())
            .then(data => {
                if (data.Status > 0) {
                    if (data.Status === 1) {
                        commandQueueRunning.className = 'command-queue-done';
                        commandQueueStatusText.innerText = `"${text}" Завершена!`;
                        commandQueuePrompt.innerText = data.Output;
                        commandQueueStatusSymbol.src = `${staticUrl}/media/vector/check.svg`;
                    }
                    else {
                        commandQueueRunning.className = 'command-queue-failed';
                        commandQueueStatusText.innerText = 'Ошибка!';
                        commandQueuePrompt.innerText = data.Output;
                        commandQueueStatusSymbol.src = `${staticUrl}/media/vector/emoji-frown.svg`;
                    }
                    clearInterval(interval);
                }
            })
    }, 3000);
}

// Extracts command from prompt as texts
function extractCommandFromPrompt() {
    let constructed_command = '';

    for (let i = 0; i < prompt.childNodes.length; i++) {
        let command_wrapper = prompt.childNodes[i];

        let command_text = Array.from(command_wrapper.childNodes).find(
            child => child.className === 'command-prompt-text'
        ).innerText;

        let command_args_node = Array.from(command_wrapper.childNodes).find(
            child => child.className === 'command-prompt-args'
        )

        if (command_args_node) {
            command_text = command_text.replace('?', command_args_node.value);
        }

        constructed_command += command_text + ' ';
    }

    return constructed_command;
}

// Loads commands from server and places them into palette
async function populatePalette() {
    let data = await fetch(`${apiUrl}/palette/`, {
        method: "POST",
        mode: 'cors',
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            Username: modal_username.value,
            UserToken: userToken
        })
    }).then(data => data.json())

    commands = [];

    while (palette.firstChild) {
        palette.removeChild(palette.firstChild);
    }

    for (let text of data.Commands) {
        commands.push(new Command(text, true ? text.includes('?') : false));
    }

    for (let command of commands) {
        palette.appendChild(constructPaletteCommand(command.text));
    }
}

// Handler for logging in
modal_login.addEventListener('click', () => {
    const body = {
        Username: modal_username.value,
        Password: modal_password.value,
    };
    fetch(`${apiUrl}/login/`, {
        method: "POST",
        mode: 'cors',
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body)
    })
    .then(data => data.json())
    .then(data => {
        userToken = data.UserToken;

        fetch(`${apiUrl}/palette/`, {
            method: "POST",
            mode: 'cors',
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                Username: modal_username.value,
                UserToken: userToken
            })
        })
        .then(data => data.json())
        .then(data => {
            for (let text of data.Commands) {
                commands.push(new Command(text, true ? text.includes('?') : false));
            }

            for (let command of commands) {
                palette.appendChild(constructPaletteCommand(command.text));
            }
        })
        .catch(err => { console.log(err); });

        modal.classList.add('hidden');
    })
    .catch(err => { console.log(err); });
});

// Handler for executing commands from prompt
execute.addEventListener('click', () => {
    const command = extractCommandFromPrompt();

    const body = {
        Text: command,
        UserToken: userToken
    }

    fetch(`${apiUrl}/command/execute/`, {
        method: "POST",
        mode: 'cors',
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body)
    })
        .then(data => data.json())
        .then(data => {
            appendTaskRunning(data.TaskToken, command)
        }).catch(data => {
    });
});

// Handler for saving commands
save.addEventListener('click', () => {
    const command = extractCommandFromPrompt();

    const body = {
        Command: command,
        Username: modal_username.value,
        UserToken: userToken
    }

    fetch(`${apiUrl}/command/save/`, {
        method: "POST",
        mode: 'cors',
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body)
    })
    .then(() => {
        setTimeout(() => {
            populatePalette()
        }, 300);
    })
    .catch(err => { 
        console.log(err);
    });
});
