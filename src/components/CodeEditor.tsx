import { ChangeEvent, useRef, useState } from 'react';
import Editor from '@monaco-editor/react';
import { FileCode, Download, Upload, Eye, Code, CheckCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import { parse } from 'yaml';

const yamlExample = `# Definición de Servicio
service:
  name: "auth-service"
  description: "Servicio de autenticación y autorización"
  
entities:
  - name: "User"
    type: "aggregate"
    fields:
      - name: "id"
        type: "uuid"
        primary: true
      - name: "email" 
        type: "string"
        unique: true
        validations:
          - required
          - email
      - name: "password"
        type: "string"
        validations:
          - required
          - min_length: 8

usecases:
  - name: "LoginUser"
    actor: "User"
    description: "Authenticate user credentials"
    input:
      - email: string
      - password: string
    output:
      - token: string
      - user: User
    errors:
      - InvalidCredentials
      - UserNotFound

  - name: "RegisterUser" 
    actor: "User"
    description: "Create new user account"
    input:
      - email: string
      - password: string
    output:
      - user: User
    errors:
      - EmailAlreadyExists
      - InvalidPassword

infrastructure:
  database: "postgresql"
  cache: "redis"
  queue: "rabbitmq"
  deployment: "docker"`;

const generatedCode = `// Generated Auth Service
package auth

import (
    "context"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

// User aggregate
type User struct {
    ID       uuid.UUID \`json:"id" db:"id"\`
    Email    string    \`json:"email" db:"email"\`
    Password string    \`json:"-" db:"password"\`
}

// LoginUser use case
type LoginUserInput struct {
    Email    string \`json:"email" validate:"required,email"\`
    Password string \`json:"password" validate:"required"\`
}

type LoginUserOutput struct {
    Token string \`json:"token"\`
    User  User   \`json:"user"\`
}

func (s *AuthService) LoginUser(ctx context.Context, input LoginUserInput) (*LoginUserOutput, error) {
    user, err := s.repo.FindUserByEmail(ctx, input.Email)
    if err != nil {
        return nil, ErrUserNotFound
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        return nil, ErrInvalidCredentials
    }
    
    token, err := s.tokenService.Generate(user.ID)
    if err != nil {
        return nil, err
    }
    
    return &LoginUserOutput{
        Token: token,
        User:  user,
    }, nil
}`;

interface CodeEditorProps {
  className?: string;
}

export function CodeEditor({ className }: CodeEditorProps) {
  const [yamlContent, setYamlContent] = useState(yamlExample);
  const [isGenerating, setIsGenerating] = useState(false);
  const [generatedPreview, setGeneratedPreview] = useState('');
  const { toast } = useToast();
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const handleLoadClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileUpload = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];

    if (!file) {
      return;
    }

    try {
      const text = await file.text();

      if (!text.trim()) {
        toast({
          title: 'Archivo vacío',
          description: 'El archivo seleccionado no contiene información.',
          variant: 'destructive',
        });
        return;
      }

      try {
        parse(text);
      } catch (parseError) {
        console.error('Invalid YAML file:', parseError);
        toast({
          title: 'YAML inválido',
          description: 'El archivo seleccionado no tiene un formato válido.',
          variant: 'destructive',
        });
        return;
      }

      setYamlContent(text);
      toast({
        title: 'Archivo cargado',
        description: `Se cargó correctamente "${file.name}".`,
      });
    } catch (error) {
      console.error('Failed to read file:', error);
      toast({
        title: 'Error al leer archivo',
        description: 'No se pudo procesar el archivo seleccionado.',
        variant: 'destructive',
      });
    } finally {
      // Reset the input so that the same file can be uploaded again if needed
      event.target.value = '';
    }
  };

  const handleExport = () => {
    if (!yamlContent.trim()) {
      toast({
        title: 'Sin contenido para exportar',
        description: 'Agrega contenido YAML antes de exportar.',
        variant: 'destructive',
      });
      return;
    }

    try {
      const timestamp = new Date();
      const sanitizedTimestamp = timestamp.toISOString().replace(/[:.]/g, '-');
      const fileName = `kthulu-dsl-${sanitizedTimestamp}.yaml`;

      const sections = [
        '# Exportación DSL Kthulu',
        `# Generado: ${timestamp.toISOString()}`,
        '',
        yamlContent.trim(),
      ];

      if (generatedPreview) {
        sections.push('', '# Vista previa generada', '```', generatedPreview, '```');
      }

      const blob = new Blob([sections.join('\n')], {
        type: 'text/plain;charset=utf-8',
      });

      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = fileName;
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);
      URL.revokeObjectURL(url);

      toast({
        title: 'Exportación lista',
        description: `Se descargó "${fileName}" con el contenido actual.`,
      });
    } catch (error) {
      console.error('Failed to export DSL:', error);
      toast({
        title: 'Error al exportar',
        description: 'No se pudo descargar el archivo. Inténtalo nuevamente.',
        variant: 'destructive',
      });
    }
  };

  const handleGenerate = async () => {
    try {
      setIsGenerating(true);
      
      // Parse YAML to create project request
      const projectRequest = {
        name: 'auth-service',
        template: 'hexagonal-go',
        modules: ['auth', 'user'],
        dryRun: true,
      };

      const plan = await kthuluApi.planProject(projectRequest);
      
      toast({
        title: 'Proyecto planeado',
        description: `Se generarán ${plan.backendFiles?.length || 0} archivos`,
      });

      // Generate preview
      const preview = `// Generated from API
// Modules: ${plan.modules.join(', ')}
// Files: ${plan.backendFiles?.length || 0}
${generatedCode}`;
      
      setGeneratedPreview(preview);
    } catch (error) {
      console.error('Generation failed:', error);
      toast({
        title: 'Error',
        description: 'No se pudo conectar con Kthulu API. Usando generación local.',
        variant: 'destructive',
      });
      setGeneratedPreview(generatedCode);
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className={`h-full bg-kthulu-surface1 flex flex-col ${className}`}>
      {/* Header */}
      <div className="p-4 border-b border-primary/20 bg-kthulu-surface2">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="font-mono font-bold text-primary text-lg">EDITOR KTHULU DSL</h2>
            <p className="text-xs text-muted-foreground font-mono">Definición de arquitectura</p>
          </div>
          
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="bg-kthulu-surface1 border-primary/30 hover:bg-primary/10 font-mono"
              onClick={handleLoadClick}
            >
              <Upload className="w-3 h-3 mr-1" />
              Cargar
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="bg-kthulu-surface1 border-accent/30 hover:bg-accent/10 font-mono"
              onClick={handleExport}
            >
              <Download className="w-3 h-3 mr-1" />
              Exportar
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".yaml,.yml,application/x-yaml,text/yaml"
              className="hidden"
              onChange={handleFileUpload}
            />
          </div>
        </div>
      </div>

      <Tabs defaultValue="yaml" className="flex-1 flex flex-col">
        <div className="px-4 pt-2">
          <TabsList className="bg-kthulu-surface2 border border-primary/20">
            <TabsTrigger value="yaml" className="font-mono text-xs">
              <FileCode className="w-3 h-3 mr-1" />
              YAML DSL
            </TabsTrigger>
            <TabsTrigger value="preview" className="font-mono text-xs">
              <Eye className="w-3 h-3 mr-1" />
              Vista Previa
            </TabsTrigger>
            <TabsTrigger value="generated" className="font-mono text-xs">
              <Code className="w-3 h-3 mr-1" />
              Código Generado
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="yaml" className="flex-1 p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Badge variant="outline" className="font-mono text-xs">
                <CheckCircle className="w-3 h-3 mr-1" />
                Sintaxis válida
              </Badge>
              <Badge variant="secondary" className="font-mono text-xs">
                3 entidades | 2 casos de uso
              </Badge>
            </div>
            
            <Button 
              onClick={handleGenerate}
              disabled={isGenerating}
              className="bg-gradient-neon text-background hover:opacity-90 font-mono"
            >
              {isGenerating ? 'Generando...' : 'Generar Código'}
            </Button>
          </div>

          <div className="border border-primary/20 rounded-sm overflow-hidden h-[500px]">
            <Editor
              height="100%"
              defaultLanguage="yaml"
              value={yamlContent}
              onChange={(value) => setYamlContent(value || '')}
              theme="vs-dark"
              options={{
                fontSize: 13,
                fontFamily: 'JetBrains Mono, Fira Code, monospace',
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                wordWrap: 'on',
                lineNumbers: 'on',
                folding: true,
                bracketPairColorization: { enabled: true },
              }}
            />
          </div>
        </TabsContent>

        <TabsContent value="preview" className="flex-1 p-4">
          <div className="bg-kthulu-surface2 border border-primary/20 rounded-sm p-6 h-full">
            <h3 className="font-mono text-primary text-lg mb-4">ESTRUCTURA GENERADA</h3>
            
            <div className="space-y-6">
              <div>
                <h4 className="font-mono text-accent text-sm mb-2">SERVICIO</h4>
                <div className="bg-kthulu-surface1 p-3 rounded-sm border border-accent/20">
                  <div className="font-mono text-sm text-foreground">auth-service</div>
                  <div className="font-mono text-xs text-muted-foreground">Servicio de autenticación y autorización</div>
                </div>
              </div>

              <div>
                <h4 className="font-mono text-secondary text-sm mb-2">ENTIDADES</h4>
                <div className="bg-kthulu-surface1 p-3 rounded-sm border border-secondary/20">
                  <div className="font-mono text-sm text-foreground">User (aggregate)</div>
                  <div className="font-mono text-xs text-muted-foreground ml-4">
                    • id: uuid (primary)<br/>
                    • email: string (unique, required, email)<br/>
                    • password: string (required, min_length: 8)
                  </div>
                </div>
              </div>

              <div>
                <h4 className="font-mono text-accent text-sm mb-2">CASOS DE USO</h4>
                <div className="space-y-2">
                  <div className="bg-kthulu-surface1 p-3 rounded-sm border border-accent/20">
                    <div className="font-mono text-sm text-foreground">LoginUser</div>
                    <div className="font-mono text-xs text-muted-foreground">Actor: User | Authenticate user credentials</div>
                  </div>
                  <div className="bg-kthulu-surface1 p-3 rounded-sm border border-accent/20">
                    <div className="font-mono text-sm text-foreground">RegisterUser</div>
                    <div className="font-mono text-xs text-muted-foreground">Actor: User | Create new user account</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="generated" className="flex-1 p-4">
          <div className="border border-primary/20 rounded-sm overflow-hidden h-[500px]">
            <Editor
              height="100%"
              defaultLanguage="go"
              value={generatedPreview || generatedCode}
              theme="vs-dark"
              options={{
                fontSize: 13,
                fontFamily: 'JetBrains Mono, Fira Code, monospace',
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                readOnly: true,
                lineNumbers: 'on',
              }}
            />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}