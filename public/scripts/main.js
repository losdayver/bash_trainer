const palette = document.getElementById('palette');
const prompt = document.getElementById('prompt');
const queue = document.getElementById('queue');

const execute = document.getElementById('prompt-buttons-execute');
const save = document.getElementById('prompt-buttons-save');

const api_hostname = 'http://localhost:4000'

function onDragStart(event) {
    event
        .dataTransfer
        .setData('text/plain', event.target.id);
}

function onDragOver(event) {
    event.preventDefault();

    const dropzone = event.target;

    if (dropzone.classList.contains('command-prompt')) {
        dropzone.classList.add('command-prompt-draggedover');
    }
}

function onDragLeave(event) {
    event.preventDefault();

    const dropzone = event.target;

    if (dropzone.classList.contains('command-prompt')) {
        dropzone.classList.remove('command-prompt-draggedover');
    }
}

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

function appendTaskRunning(token, text) {
    const commandQueueRunning = document.createElement('div');
    commandQueueRunning.className = 'command-queue-running';
    commandQueueRunning.data = token;

    const commandQueueStatus = document.createElement('div');
    commandQueueStatus.classList.add('command-queue-status');

    const commandQueueStatusSymbol = document.createElement('img');
    commandQueueStatusSymbol.src = '../media/vector/three-dots.svg';
    commandQueueStatusSymbol.classList.add('command-queue-status-symbol');

    const commandQueuePrompt = document.createElement('div');
    commandQueuePrompt.classList.add('command-queue-prompt');
    commandQueuePrompt.innerText = text

    const commandQueueStatusText = document.createElement('p');
    commandQueueStatusText.classList.add('command-queue-status-text');
    commandQueueStatusText.innerText = 'Running...';

    const commandQueueCross = document.createElement('img');
    commandQueueCross.src = '../media/vector/x.svg';
    commandQueueCross.classList.add('command-queue-cross');

    commandQueueRunning.appendChild(commandQueueStatus);
    commandQueueStatus.appendChild(commandQueueStatusSymbol);
    commandQueueStatus.appendChild(commandQueueStatusText);

    commandQueueRunning.appendChild(commandQueuePrompt);
    commandQueueRunning.appendChild(commandQueueCross);

    queue.appendChild(commandQueueRunning);

    commandQueueCross.addEventListener('click', () => {
        if (commandQueueRunning.className != 'command-queue-running') {
            remove(commandQueueRunning);
        }
    })

    let interval;
    interval = setInterval(() => {
        fetch(api_hostname + '/api/task/' + token)
            .then(data => data.json())
            .then(data => {
                if (data.Status > 0) {
                    if (data.Status === 1) {
                        commandQueueRunning.className = 'command-queue-done';
                        commandQueueStatusText.innerText = `"${text}" Done!`;
                        commandQueuePrompt.innerText = data.Output;
                        commandQueueStatusSymbol.src = '../media/vector/check.svg';
                    }
                    else {
                        commandQueueRunning.className = 'command-queue-failed';
                        commandQueueStatusText.innerText = 'Failed!';
                        commandQueuePrompt.innerText = data.Output;
                        commandQueueStatusSymbol.src = '../media/vector/emoji-frown.svg';
                    }
                    clearInterval(interval);
                }
            })
    }, 3000);
}

class Command {
    constructor(text, hasArgs) {
        this.text = text;
        this.hasArgs = !!hasArgs;
    }
}

const commands = [
    new Command('cd ?;', true),
    new Command('ls -1', false),
    new Command('xargs cat', false),
    new Command('find . ? -type f', true),
    new Command('wc -c', false),
    new Command('wc -w', false),
    new Command('wc -l', false),
    new Command('grep ?', true),
    new Command('sed -n "?p"', true)
]

for (let command of commands) {
    palette.appendChild(constructPaletteCommand(command.text));
}

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

        constructed_command += command_text;

        if (i != prompt.childNodes.length - 1 &&
            command_text[command_text.length - 1] != ';') {
            constructed_command += ' | ';
        }
    }

    return constructed_command;
}

execute.addEventListener('click', () => {
    const command = extractCommandFromPrompt();

    const body = {
        Text: command,
        UserToken: 'testtesttest'
    }

    fetch(api_hostname + '/api/command/execute/', {
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