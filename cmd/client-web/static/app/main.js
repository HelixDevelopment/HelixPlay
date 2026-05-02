import { AppModule } from './app.module.js';

document.addEventListener('DOMContentLoaded', () => {
  const app = new AppModule();
  app.bootstrap();
});
