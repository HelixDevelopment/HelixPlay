export class StreamingComponent {
  constructor() {
    this.canvas = document.getElementById('streamCanvas');
    this.ctx = this.canvas.getContext('2d');
    this.animationId = null;
  }

  init() {
    this.resize();
    window.addEventListener('resize', () => this.resize());
    this.startRenderLoop();
    this.startInputLoop();
  }

  resize() {
    this.canvas.width = this.canvas.clientWidth;
    this.canvas.height = this.canvas.clientHeight;
  }

  startRenderLoop() {
    const render = () => {
      this.ctx.fillStyle = '#000';
      this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
      this.ctx.fillStyle = '#0f0';
      this.ctx.font = '24px sans-serif';
      this.ctx.fillText('Streaming Canvas', 20, 40);
      this.animationId = requestAnimationFrame(render);
    };
    render();
  }

  startInputLoop() {
    setInterval(() => {
      const pads = navigator.getGamepads ? navigator.getGamepads() : [];
      for (const pad of pads) {
        if (!pad) continue;
        const state = {
          id: pad.index,
          connected: pad.connected,
          buttonMask: this.buttonsToMask(pad.buttons),
          leftStickX: pad.axes[0] ? Math.round(pad.axes[0] * 32767) : 0,
          leftStickY: pad.axes[1] ? Math.round(pad.axes[1] * 32767) : 0,
          rightStickX: pad.axes[2] ? Math.round(pad.axes[2] * 32767) : 0,
          rightStickY: pad.axes[3] ? Math.round(pad.axes[3] * 32767) : 0,
        };
        fetch('/api/streaming/controller', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(state),
        }).catch(() => {});
      }
    }, 16);
  }

  buttonsToMask(buttons) {
    let mask = 0;
    for (let i = 0; i < buttons.length && i < 32; i++) {
      if (buttons[i].pressed) mask |= (1 << i);
    }
    return mask;
  }
}
