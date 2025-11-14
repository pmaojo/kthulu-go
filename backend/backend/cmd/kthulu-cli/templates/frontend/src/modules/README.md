# Módulos de frontend

Los módulos pueden registrar rutas y componentes reutilizables a través de `registerModule`.

## Componentes de layout

`AdminLayout` acepta tres puntos de extensión: `header`, `sidebar` y `footer`. Un módulo puede suministrar sus propios elementos registrándolos en la clave `components`:

```ts
import { registerModule } from '../registry';
import CustomHeader from './CustomHeader';

registerModule({
  routes: [],
  components: {
    header: CustomHeader,
    sidebar: CustomSidebar,
    footer: CustomFooter,
  },
});
```

Cuando se crea el router, estos componentes se pasan a `AdminLayout` como props. Si no se registra ninguno, se renderizan placeholders por defecto.
