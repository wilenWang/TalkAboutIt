export interface PersonaLanguage {
  primary: string;
  allowed: string[];
  default_output: string;
  style_hint: string;
}

export interface PersonaStance {
  default_position: string;
  intensity: number;
  biases: string[];
  taboos: string[];
}

export interface CoreBelief {
  belief: string;
  priority: number;
  rationale: string;
}

export interface SpeakingStyle {
  tone: string;
  cadence: string;
  verbosity: number;
  signature_patterns: string[];
  do: string[];
  dont: string[];
}

export interface KnowledgeScope {
  domains: string[];
  expertise_level: Record<string, number>;
  time_cutoff: string;
  allowed_inference: string;
  unknown_handling: string;
  forbidden_claims: string[];
}

export interface InteractionRules {
  address_others: string;
  disagreement_style: string;
  interruption_policy: string;
  question_policy: string;
  concession_policy: string;
  avoid: string[];
}

export interface DebateGoal {
  primary_goal: string;
  secondary_goals: string[];
  win_condition: string;
  loss_condition: string;
}

export interface Prompting {
  system_preamble: string;
  reply_constraints: string[];
}

export interface Examples {
  opening_line: string;
  sample_rebuttal: string;
}

export interface Persona {
  schema_version: string;
  id: string;
  name: string;
  display_name: string;
  avatar: string;
  role_title: string;
  description: string;
  tags: string[];
  archetype: string;
  language: PersonaLanguage;
  stance: PersonaStance;
  core_beliefs: CoreBelief[];
  speaking_style: SpeakingStyle;
  knowledge_scope: KnowledgeScope;
  interaction_rules: InteractionRules;
  debate_goal: DebateGoal;
  prompting: Prompting;
  examples: Examples;
}

export function createEmptyPersona(): Persona {
  return {
    schema_version: 'persona.v1',
    id: '',
    name: '',
    display_name: '',
    avatar: '🤖',
    role_title: '',
    description: '',
    tags: [],
    archetype: '',
    language: {
      primary: 'zh-CN',
      allowed: ['zh-CN', 'en-US'],
      default_output: 'follow_user',
      style_hint: '',
    },
    stance: {
      default_position: '',
      intensity: 3,
      biases: [],
      taboos: [],
    },
    core_beliefs: [],
    speaking_style: {
      tone: '',
      cadence: '',
      verbosity: 3,
      signature_patterns: [],
      do: [],
      dont: [],
    },
    knowledge_scope: {
      domains: [],
      expertise_level: {},
      time_cutoff: '',
      allowed_inference: '',
      unknown_handling: '',
      forbidden_claims: [],
    },
    interaction_rules: {
      address_others: '',
      disagreement_style: '',
      interruption_policy: '',
      question_policy: '',
      concession_policy: '',
      avoid: [],
    },
    debate_goal: {
      primary_goal: '',
      secondary_goals: [],
      win_condition: '',
      loss_condition: '',
    },
    prompting: {
      system_preamble: '',
      reply_constraints: [],
    },
    examples: {
      opening_line: '',
      sample_rebuttal: '',
    },
  };
}
