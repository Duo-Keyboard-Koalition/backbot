declare module 'qrcode-terminal' {
  export function generate(text: string, options?: { small?: boolean }): void;
}

export interface AgentRegisterRequest {
  agentName: string;
  agentType: 'scorpion' | 'nanobot' | 'openclaw' | 'zeroclaw' | 'custom';
  publicKeyPem: string;
  metadata?: Record<string, unknown>;
}

export interface AgentRegisterResponse {
  agentId: string;
  agentName: string;
  publicKeyFingerprint: string;
}

export interface AgentLoginChallengeRequest {
  agentId: string;
}

export interface AgentLoginChallengeResponse {
  challengeId: string;
  challenge: string;
  expiresAt: string;
}

export interface AgentLoginVerifyRequest {
  agentId: string;
  challengeId: string;
  signature: string;
}

export interface AgentSessionTokens {
  accessToken: string;
  refreshToken: string;
  expiresInSeconds: number;
}

export interface MessageOrigin {
  workspaceId: string;
  serverId: string;
  channelId: string;
  threadId?: string;
  messageId: string;
  userId: string;
}

export interface BridgeInboundMessage {
  type: 'user.message';
  text: string;
  origin: MessageOrigin;
  target: {
    agentType: string;
    agentId: string;
  };
  correlationId: string;
}

export interface BridgeOutboundMessage {
  type: 'agent.message';
  text: string;
  origin: MessageOrigin;
  bridge: {
    bridgeMsgId: string;
    correlationId: string;
    adapter: string;
    agentId: string;
  };
}
