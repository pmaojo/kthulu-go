import type { ComponentType } from 'react';
import type { RouteObject } from 'react-router-dom';

export interface Module {
  routes: RouteObject[];
  components?: Record<string, ComponentType>;
}

const registeredModules: Module[] = [];

export function registerModule(module: Module) {
  registeredModules.push(module);
}

export function getRegisteredRoutes(): RouteObject[] {
  return registeredModules.flatMap((m) => m.routes);
}

export function getRegisteredComponents(): Record<string, ComponentType> {
  return registeredModules.reduce<Record<string, ComponentType>>((acc, m) => {
    if (m.components) {
      Object.assign(acc, m.components);
    }
    return acc;
  }, {});
}
