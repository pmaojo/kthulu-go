import React, { useState, useEffect } from 'react';
import { kthuluApi } from '../services/kthuluApi';
import { Card, Layout, Typography, List, Tag, Button, Split, message } from 'antd';
import { PlayCircleOutlined, FileTextOutlined } from '@ant-design/icons';

const { Content, Sider } = Layout;
const { Title, Text, Paragraph } = Typography;

interface FeatureFile {
  path: string;
  name: string;
  scenarios: string[]; // Mocked for now
}

const BehaviorLab: React.FC = () => {
  const [features, setFeatures] = useState<FeatureFile[]>([]);
  const [selectedFeature, setSelectedFeature] = useState<FeatureFile | null>(null);
  const [testResults, setTestResults] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadFeatures();
  }, []);

  const loadFeatures = async () => {
    try {
      // In a real implementation, we would parse the output of listFeatures
      // For now, we mock the parsing or expect the CLI to return JSON
      const result = await kthuluApi.listFeatures();
      if (result.output && result.output.length > 0) {
        // Assume output is list of paths
        const paths = result.output[0].split('\n').filter(Boolean);
        const featureList = paths.map(p => ({
          path: p,
          name: p.split('/').pop() || p,
          scenarios: ['Scenario 1', 'Scenario 2'] // detailed parsing needed
        }));
        setFeatures(featureList);
      }
    } catch (e) {
      console.error("Failed to load features", e);
      message.error("Failed to load features");
    }
  };

  const runTests = async (filter: string = '') => {
    setLoading(true);
    setTestResults('');
    try {
      const result = await kthuluApi.runScenario(filter);
      setTestResults(result.output.join('\n'));
      if (result.status !== 'success') {
          message.warning("Tests failed (Red)");
      } else {
          message.success("Tests passed (Green)");
      }
    } catch (e) {
      message.error("Failed to run tests");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout style={{ height: '100vh' }}>
      <Sider width={300} theme="light" style={{ borderRight: '1px solid #f0f0f0' }}>
        <div style={{ padding: 16 }}>
          <Title level={4}><FileTextOutlined /> Features</Title>
          <Button onClick={loadFeatures} block>Refresh</Button>
        </div>
        <List
          dataSource={features}
          renderItem={item => (
            <List.Item
                onClick={() => setSelectedFeature(item)}
                style={{ cursor: 'pointer', background: selectedFeature?.path === item.path ? '#e6f7ff' : 'transparent', padding: '10px 16px' }}
            >
              <Text>{item.name}</Text>
            </List.Item>
          )}
        />
      </Sider>
      <Content style={{ padding: 24, display: 'flex', flexDirection: 'column' }}>
        {selectedFeature ? (
            <>
                <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Title level={3}>{selectedFeature.name}</Title>
                    <Button
                        type="primary"
                        icon={<PlayCircleOutlined />}
                        loading={loading}
                        onClick={() => runTests(selectedFeature.path)}
                    >
                        Run Feature
                    </Button>
                </div>
                <div style={{ display: 'flex', flex: 1, gap: 16 }}>
                    <Card title="Feature Spec" style={{ flex: 1, overflow: 'auto' }}>
                        <Paragraph>
                            <pre>{`Feature: ${selectedFeature.name}
  Scenario: Example
    Given ...
    When ...
    Then ...`}</pre>
                            {/* In real app, we fetch file content here */}
                        </Paragraph>
                    </Card>
                    <Card title="Test Output" style={{ flex: 1, overflow: 'auto' }}>
                         <pre style={{
                             background: '#000',
                             color: '#0f0',
                             padding: 12,
                             borderRadius: 4,
                             height: '100%',
                             whiteSpace: 'pre-wrap'
                         }}>
                             {testResults || "Run tests to see output..."}
                         </pre>
                    </Card>
                </div>
            </>
        ) : (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                <Text type="secondary">Select a feature to view scenarios</Text>
            </div>
        )}
      </Content>
    </Layout>
  );
};

export default BehaviorLab;
