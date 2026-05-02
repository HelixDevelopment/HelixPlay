export class CatalogComponent {
  constructor() {
    this.listEl = document.getElementById('gameList');
  }

  async init() {
    try {
      const resp = await fetch('/api/catalog/games?tenant_id=demo');
      if (!resp.ok) throw new Error('Failed to load catalog');
      const data = await resp.json();
      this.render(data.games || []);
    } catch (e) {
      this.listEl.innerHTML = '<li>Offline catalog</li>';
    }
  }

  render(games) {
    this.listEl.innerHTML = '';
    for (const game of games) {
      const li = document.createElement('li');
      li.textContent = game.title || game.game_id;
      this.listEl.appendChild(li);
    }
  }
}
