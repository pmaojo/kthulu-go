export type ElementType = "service" | "entity" | "usecase";

export interface ElementProperties {
  id: string;
  type: ElementType;
  name: string;
  description: string;
  fields?: string[];
  actor?: string;
  action?: string;
  status?: "active" | "inactive" | "error";
}
