import {Constants} from '../constants';
import {showMessage} from '../util/message';

export class WebSocketUtil {
    public webSocket: WebSocket;

    constructor(liandi: ILiandi) {
        this.connect(liandi);
    }

    private connect(liandi: ILiandi) {
        this.webSocket = new WebSocket(Constants.WEBSOCKET_ADDREDD);
        this.webSocket.onopen = () => {
            liandi.ws.webSocket.send(JSON.stringify({
                cmd: 'dirs',
                param: {},
            }));
        };
        this.webSocket.onclose = (e) => {
            console.warn('WebSocket is closed. Reconnect will be attempted in 1 second.', e.reason);
            setTimeout(() => {
                this.connect(liandi);
            }, 1000);
        };
        this.webSocket.onerror = (err) => {
            console.error('WebSocket Error:', err);
            this.webSocket.close();
        };
        this.webSocket.onmessage = (event) => {
            const response = JSON.parse(event.data);
            if (response.code !== 0) {
                showMessage(response.msg, 0);
                return;
            }
            switch (response.cmd) {
                case 'mount':
                    liandi.navigation.onMount(liandi, response.data.url);
                    break;
                case 'ls':
                    liandi.files.onLs(liandi, response.data);
                    break;
                case 'get':
                    liandi.editors.onGet(liandi, response.data);
                    break;
                case 'dirs':
                    liandi.navigation.element.lastElementChild.innerHTML = '';
                    response.data.forEach((url: string) => {
                        liandi.navigation.onMount(liandi, url);
                    });
                    break;
            }
        };
    }
}
