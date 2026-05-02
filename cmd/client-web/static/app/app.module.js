import { StreamingComponent } from './components/streaming/streaming.component.js';
import { CatalogComponent } from './components/catalog/catalog.component.js';

export class AppModule {
  constructor() {
    this.streaming = new StreamingComponent();
    this.catalog = new CatalogComponent();
  }

  bootstrap() {
    this.streaming.init();
    this.catalog.init();
  }
}
