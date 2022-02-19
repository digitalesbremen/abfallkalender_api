import {css, html, LitElement} from "lit";

class Kalender extends LitElement {

    render() {
        return html`
            <div class="test-class">
                Hallo Bremer Abfallkalender API!
            </div>`;
    }

    static get styles() {
        return css`
            :host .test-class {
                background-color: blue;
                color: white;
            }`;
    }
}

customElements.define('kalender-component', Kalender)