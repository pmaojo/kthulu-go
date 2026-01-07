import { useState, useEffect, useCallback } from 'react';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';

export interface FeatureFile {
  path: string;
  name: string;
  scenarios: string[];
}

export function useBehaviorLab() {
  const [features, setFeatures] = useState<FeatureFile[]>([]);
  const [selectedFeature, setSelectedFeature] = useState<FeatureFile | null>(null);
  const [testResults, setTestResults] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const { toast } = useToast();

  const loadFeatures = useCallback(async () => {
    try {
      setIsLoading(true);
      const result = await kthuluApi.listFeatures();

      if (result.output && result.output.length > 0) {
        const rawOutput = result.output.join('\n');
        const paths = rawOutput.split('\n')
            .map(s => s.trim())
            .filter(s => s.endsWith('.feature'));

        const featureList = paths.map(p => ({
          path: p,
          name: p.split('/').pop() || p,
          scenarios: ['Scenario 1', 'Scenario 2']
        }));
        setFeatures(featureList);
      }
    } catch (e) {
      console.error("Failed to load features", e);
      toast({
        title: "Error",
        description: "No se pudieron cargar los archivos de features.",
        variant: "destructive",
      });
      // Fallback mock
      setFeatures([
        { path: 'features/auth.feature', name: 'auth.feature', scenarios: ['Login', 'Logout'] },
        { path: 'features/products.feature', name: 'products.feature', scenarios: ['List Products', 'Get Product'] },
      ]);
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    loadFeatures();
  }, [loadFeatures]);

  const runTests = async (filter: string = '') => {
    setIsRunning(true);
    setTestResults('Ejecutando tests...');
    try {
      const result = await kthuluApi.runScenario(filter);
      setTestResults(result.output.join('\n'));

      if (result.status === 'success' || (result.metadata?.exitCode === 0)) {
        toast({
          title: "Tests Pasaron",
          description: "La ejecución de features fue exitosa.",
          variant: "default",
          className: "bg-green-500/10 border-green-500/50 text-green-500",
        });
      } else {
        toast({
          title: "Tests Fallaron",
          description: "Hubo errores en la ejecución de features.",
          variant: "destructive",
        });
      }
    } catch (e) {
      setTestResults('Error al ejecutar tests.');
      toast({
        title: "Error",
        description: "Fallo al invocar el comando de tests.",
        variant: "destructive",
      });
    } finally {
      setIsRunning(false);
    }
  };

  return {
    features,
    selectedFeature,
    setSelectedFeature,
    testResults,
    isLoading,
    isRunning,
    loadFeatures,
    runTests
  };
}
