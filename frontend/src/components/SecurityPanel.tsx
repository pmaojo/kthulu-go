import { useEffect, useMemo, useState } from 'react';
import { Shield, Loader2, RefreshCcw } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type { SecurityConfig } from '@/types/kthulu';

const logLevels = ['debug', 'info', 'warn', 'error'] as const;
const sameSiteModes = ['lax', 'strict', 'none'] as const;

export function SecurityPanel() {
  const { toast } = useToast();
  const [config, setConfig] = useState<SecurityConfig | null>(null);
  const [draft, setDraft] = useState<SecurityConfig | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);

  const loadConfig = async () => {
    setLoading(true);
    try {
      const response = await kthuluApi.getSecurityConfig();
      setConfig(response);
      setDraft(response);
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'No se pudo cargar la configuración.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfig();
  }, []);

  const updateDraftSection = <K extends keyof SecurityConfig>(section: K, values: Partial<NonNullable<SecurityConfig[K]>>) => {
    setDraft((previous) => ({
      ...(previous ?? {}),
      [section]: {
        ...((previous && previous[section]) ?? {}),
        ...values,
      },
    }));
  };

  const validateDraft = (nextDraft: SecurityConfig | null) => {
    if (!nextDraft) return 'No hay configuración para actualizar.';
    if (nextDraft.session?.max_age !== undefined && nextDraft.session.max_age < 0) {
      return 'La vida de la sesión debe ser positiva.';
    }
    if (nextDraft.rbac?.cache_enabled && !nextDraft.rbac.cache_ttl) {
      return 'Define un TTL de cache para RBAC cuando el cache está activo.';
    }
    return null;
  };

  const handleSave = async () => {
    const validation = validateDraft(draft);
    if (validation) {
      setValidationError(validation);
      toast({
        title: 'Configuración inválida',
        description: validation,
        variant: 'destructive',
      });
      return;
    }

    if (!draft) return;
    setValidationError(null);
    setSaving(true);
    const previous = config;
    setConfig(draft);

    try {
      const updated = await kthuluApi.updateSecurityConfig(draft);
      setConfig(updated);
      setDraft(updated);
      toast({
        title: 'Configuración guardada',
        description: 'Los cambios de seguridad fueron aplicados.',
      });
    } catch (err) {
      setConfig(previous ?? null);
      setDraft(previous ?? null);
      const message = err instanceof Error ? err.message : 'No se pudieron guardar los cambios.';
      toast({
        title: 'Error guardando seguridad',
        description: message,
        variant: 'destructive',
      });
    } finally {
      setSaving(false);
    }
  };

  const dirty = useMemo(() => JSON.stringify(config) !== JSON.stringify(draft), [config, draft]);

  return (
    <div className="h-full bg-kthulu-surface1 p-6 overflow-y-auto">
      <div className="max-w-5xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-mono text-primary font-bold flex items-center gap-2">
              <Shield className="w-6 h-6" /> Seguridad y Políticas
            </h2>
            <p className="text-sm text-muted-foreground font-mono">
              Administra RBAC, auditoría y sesiones respaldadas por el API de Kthulu.
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={loadConfig} disabled={loading || saving}>
              <RefreshCcw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
              Recargar
            </Button>
            <Button onClick={handleSave} disabled={saving || loading || !dirty || !draft}>
              {saving && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
              Guardar cambios
            </Button>
          </div>
        </div>

        {error && (
          <div className="p-3 border border-destructive/40 bg-destructive/10 text-destructive font-mono text-sm rounded">
            {error}
          </div>
        )}
        {validationError && (
          <div className="p-3 border border-accent/40 bg-accent/10 text-accent font-mono text-sm rounded">
            {validationError}
          </div>
        )}

        {loading || !draft ? (
          <div className="flex items-center justify-center py-10 text-muted-foreground">
            <Loader2 className="w-5 h-5 animate-spin" />
          </div>
        ) : (
          <div className="space-y-4">
            <Card className="bg-kthulu-surface2 border-primary/20">
              <CardHeader>
                <CardTitle className="font-mono text-primary text-sm">RBAC</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-mono text-sm">RBAC habilitado</p>
                    <p className="text-xs text-muted-foreground">Activa políticas derivadas de tags @kthulu:security.</p>
                  </div>
                  <Switch
                    checked={!!draft.rbac?.enabled}
                    onCheckedChange={(checked) => updateDraftSection('rbac', { enabled: checked })}
                    aria-label="RBAC habilitado"
                  />
                </div>

                <div className="grid md:grid-cols-2 gap-4">
                  <ToggleRow
                    label="Modo estricto"
                    description="Solo se permiten acciones declaradas"
                    checked={!!draft.rbac?.strict_mode}
                    onChange={(checked) => updateDraftSection('rbac', { strict_mode: checked })}
                  />
                  <ToggleRow
                    label="Política deny por defecto"
                    description="Bloquea cualquier rol no definido"
                    checked={!!draft.rbac?.default_deny_policy}
                    onChange={(checked) => updateDraftSection('rbac', { default_deny_policy: checked })}
                  />
                  <ToggleRow
                    label="Seguridad contextual"
                    description="Evalúa atributos dinámicos"
                    checked={!!draft.rbac?.contextual_security}
                    onChange={(checked) => updateDraftSection('rbac', { contextual_security: checked })}
                  />
                  <ToggleRow
                    label="Roles jerárquicos"
                    description="Permite herencia de permisos"
                    checked={!!draft.rbac?.hierarchical_roles}
                    onChange={(checked) => updateDraftSection('rbac', { hierarchical_roles: checked })}
                  />
                  <ToggleRow
                    label="Cache de decisiones"
                    description="Reduce llamadas repetidas"
                    checked={!!draft.rbac?.cache_enabled}
                    onChange={(checked) => updateDraftSection('rbac', { cache_enabled: checked })}
                  />
                  <ToggleRow
                    label="Auditoría integrada"
                    description="Registra cada decisión"
                    checked={!!draft.rbac?.audit_enabled}
                    onChange={(checked) => updateDraftSection('rbac', { audit_enabled: checked })}
                  />
                </div>

                <div className="space-y-2">
                  <Label className="font-mono text-xs text-primary" htmlFor="rbac-cache-ttl">TTL del cache</Label>
                  <Input
                    id="rbac-cache-ttl"
                    value={draft.rbac?.cache_ttl ?? ''}
                    onChange={(event) => updateDraftSection('rbac', { cache_ttl: event.target.value })}
                    placeholder="15m"
                    className="bg-kthulu-surface1 border-primary/30 font-mono"
                  />
                </div>
              </CardContent>
            </Card>

            <div className="grid md:grid-cols-2 gap-4">
              <Card className="bg-kthulu-surface2 border-primary/20">
                <CardHeader>
                  <CardTitle className="font-mono text-primary text-sm">Auditoría</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <ToggleRow
                    label="Auditoría activa"
                    description="Registro de métricas y hallazgos"
                    checked={!!draft.audit?.enabled}
                    onChange={(checked) => updateDraftSection('audit', { enabled: checked })}
                  />

                  <div className="space-y-2">
                    <Label className="font-mono text-xs text-primary" htmlFor="audit-log-level">Nivel de log</Label>
                    <Select
                      value={draft.audit?.log_level ?? 'info'}
                      onValueChange={(value) => updateDraftSection('audit', { log_level: value })}
                    >
                      <SelectTrigger id="audit-log-level" className="bg-kthulu-surface1 border-primary/30 font-mono text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent className="bg-kthulu-surface2 border-primary/20">
                        {logLevels.map((level) => (
                          <SelectItem key={level} value={level} className="font-mono text-xs">
                            {level.toUpperCase()}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label className="font-mono text-xs text-primary" htmlFor="audit-retention">Retención (días)</Label>
                    <Input
                      id="audit-retention"
                      type="number"
                      value={draft.audit?.retention_days ?? 30}
                      onChange={(event) => updateDraftSection('audit', { retention_days: Number(event.target.value) })}
                      className="bg-kthulu-surface1 border-primary/30 font-mono"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label className="font-mono text-xs text-primary" htmlFor="audit-storage">Destino</Label>
                    <Input
                      id="audit-storage"
                      value={draft.audit?.storage_type ?? ''}
                      onChange={(event) => updateDraftSection('audit', { storage_type: event.target.value })}
                      placeholder="s3, gcs, loki..."
                      className="bg-kthulu-surface1 border-primary/30 font-mono"
                    />
                  </div>
                </CardContent>
              </Card>

              <Card className="bg-kthulu-surface2 border-primary/20">
                <CardHeader>
                  <CardTitle className="font-mono text-primary text-sm">Sesiones</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <ToggleRow
                    label="Cookies seguras"
                    description="Marca las cookies como secure"
                    checked={!!draft.session?.secure_cookie}
                    onChange={(checked) => updateDraftSection('session', { secure_cookie: checked })}
                  />

                  <div className="space-y-2">
                    <Label className="font-mono text-xs text-primary" htmlFor="session-samesite">SameSite</Label>
                    <Select
                      value={draft.session?.same_site ?? 'lax'}
                      onValueChange={(value) => updateDraftSection('session', { same_site: value })}
                    >
                      <SelectTrigger id="session-samesite" className="bg-kthulu-surface1 border-primary/30 font-mono text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent className="bg-kthulu-surface2 border-primary/20">
                        {sameSiteModes.map((mode) => (
                          <SelectItem key={mode} value={mode} className="font-mono text-xs">
                            {mode.toUpperCase()}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label className="font-mono text-xs text-primary" htmlFor="session-max-age">Duración máxima (segundos)</Label>
                    <Input
                      id="session-max-age"
                      type="number"
                      value={draft.session?.max_age ?? ''}
                      onChange={(event) => updateDraftSection('session', { max_age: Number(event.target.value) })}
                      placeholder="3600"
                      className="bg-kthulu-surface1 border-primary/30 font-mono"
                    />
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

interface ToggleRowProps {
  label: string;
  description: string;
  checked: boolean;
  onChange: (value: boolean) => void;
}

function ToggleRow({ label, description, checked, onChange }: ToggleRowProps) {
  return (
    <div className="flex items-center justify-between border border-primary/20 rounded-sm px-3 py-2">
      <div>
        <p className="font-mono text-xs text-primary">{label}</p>
        <p className="text-[11px] text-muted-foreground font-mono">{description}</p>
      </div>
      <Switch checked={checked} onCheckedChange={onChange} aria-label={label} />
    </div>
  );
}
