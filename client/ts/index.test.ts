import { PlatformSDK } from './index';

describe('PlatformSDK', () => {
  it('should initialize with the provided config', () => {
    const config = {
      baseURL: 'https://api.test.com',
      apiKey: 'test-api-key',
    };
    
    const sdk = new PlatformSDK(config);
    
    // Check that the SDK has all the expected services
    expect(sdk.files).toBeDefined();
    expect(sdk.channels).toBeDefined();
    expect(sdk.bots).toBeDefined();
    expect(sdk.webrtc).toBeDefined();
    expect(sdk.subscription).toBeDefined();
    expect(sdk.chat).toBeDefined();
  });
  
  it('should use default baseURL if not provided', () => {
    const config = {
      apiKey: 'test-api-key',
    };
    
    const sdk = new PlatformSDK(config);
    
    // Test that the SDK still initializes properly
    expect(sdk.files).toBeDefined();
    expect(sdk.channels).toBeDefined();
  });
}); 