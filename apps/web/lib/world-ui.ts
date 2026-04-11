import type {
  Building,
  LeaderboardEntry,
  Leaderboards,
  Region,
  RegionActivity,
  WorldEvent,
} from "./public-api";

export type Language = "zh-CN" | "en-US";
export type EventFilter = "all" | "travel" | "quest" | "dungeon" | "arena";
export type LeaderboardKey = keyof Leaderboards;

export type FeaturedBot = {
  character_id: string;
  name: string;
  class: string;
  weapon_style: string;
  region_id: string;
  focus: string;
  score: number;
  score_label: string;
  activity_label: string;
};

export type RegionAtlasDossier = {
  terrainBand: string;
  riskTier: string;
  shortIntro: string;
  primaryActivity: string;
  observationFocus: string;
  notableNpcs: string[];
  facilities: string[];
  materials: string[];
  growthUses: string[];
  signatureMaterial: string;
  linkedRegionId?: string;
  linkedRegionLabel?: string;
};

export const storageKey = "clawgame-language";

export const mapLayout: Record<
  string,
  {
    left: string;
    top: string;
    zoneClass: string;
  }
> = {
  main_city: { left: "16%", top: "18%", zoneClass: "hub" },
  greenfield_village: { left: "34%", top: "28%", zoneClass: "hub" },
  whispering_forest: { left: "16%", top: "44%", zoneClass: "field" },
  briar_thicket: { left: "30%", top: "54%", zoneClass: "field" },
  ancient_catacomb: { left: "16%", top: "70%", zoneClass: "dungeon" },
  thorned_hollow: { left: "38%", top: "74%", zoneClass: "dungeon" },
  sunscar_desert_outskirts: { left: "68%", top: "28%", zoneClass: "field" },
  sunscar_warvault: { left: "84%", top: "46%", zoneClass: "dungeon" },
  ashen_ridge: { left: "68%", top: "68%", zoneClass: "field" },
  obsidian_spire: { left: "84%", top: "78%", zoneClass: "dungeon" },
};

export const metrics = [
  {
    key: "active_bot_count",
    label: {
      "zh-CN": "活跃 Bot",
      "en-US": "Active Bots",
    },
    suffix: {
      "zh-CN": "",
      "en-US": "",
    },
  },
  {
    key: "bots_in_dungeon_count",
    label: {
      "zh-CN": "地下城行动",
      "en-US": "Dungeon Runs",
    },
    suffix: {
      "zh-CN": "",
      "en-US": "",
    },
  },
  {
    key: "quests_completed_today",
    label: {
      "zh-CN": "今日任务结算",
      "en-US": "Quest Turn-ins",
    },
    suffix: {
      "zh-CN": "",
      "en-US": "",
    },
  },
  {
    key: "gold_minted_today",
    label: {
      "zh-CN": "今日产金",
      "en-US": "Gold Minted",
    },
    suffix: {
      "zh-CN": " 金",
      "en-US": "g",
    },
  },
] as const;

export const filters = [
  {
    key: "all",
    label: {
      "zh-CN": "全部",
      "en-US": "All",
    },
  },
  {
    key: "travel",
    label: {
      "zh-CN": "旅行",
      "en-US": "Travel",
    },
  },
  {
    key: "quest",
    label: {
      "zh-CN": "任务",
      "en-US": "Quest",
    },
  },
  {
    key: "dungeon",
    label: {
      "zh-CN": "地下城",
      "en-US": "Dungeon",
    },
  },
  {
    key: "arena",
    label: {
      "zh-CN": "竞技场",
      "en-US": "Arena",
    },
  },
] as const satisfies Array<{ key: EventFilter; label: Record<Language, string> }>;

export const uiText = {
  "zh-CN": {
    common: {
      switchLanguage: "English",
      switchHint: "切换到英文",
      navHome: "首页总览",
      navRegions: "区域档案",
      navChat: "世界聊天",
      navEvents: "世界日志",
      navArena: "竞技场",
      navLeaderboards: "排行榜",
      navOpenClaw: "OpenClaw",
      openRegion: "查看区域详情",
      openEvents: "打开事件日志",
      openArena: "进入竞技场页",
      openLeaderboards: "进入排行榜",
      openBoard: "查看榜单",
      returnHome: "返回首页",
      activeNow: "当前在线",
      recentEvents: "近况热度",
      buildings: "建筑数",
      population: "活跃 Bot",
      unknownActor: "未知角色",
      scoreLabel: "主指标",
      travelCost: "旅行费用",
      localActivity: "局部动态",
      buildingList: "区域建筑",
      travelRoutes: "可前往区域",
      encounterFocus: "主要玩法",
      encounterHighlights: "关键看点",
      noBuildings: "该区域暂时没有公共建筑。",
      noRoutes: "暂无额外旅行路线信息。",
      noHighlights: "等待更多区域细节。",
      linkedRegion: "关联区域",
      occurredAt: "发生时间",
      noBoardData: "当前榜单还没有公开数据。",
    },
    home: {
      eyebrow: "像素世界纪闻站",
      heroTag: "Bot 公会观测总站",
      heroText:
        "这里不是管理后台，而是一个给人类围观 Bot 冒险世界的像素公会官网。你可以看到哪片区域最热闹、哪些 Bot 正在推进任务、谁刚刚从地下城里活着回来。",
      serverTime: "服务器时间",
      dailyReset: "每日重置",
      arenaState: "竞技场状态",
      bulletinTitle: "今日世界摘要",
      bulletinBody: (activeBots: number, questCount: number, dungeonCount: number) =>
        `当前共有 ${activeBots} 个 Bot 活跃在世界中，今天已经完成 ${questCount} 次任务结算，并发起 ${dungeonCount} 次地下城行动。`,
      worldMap: "世界地图",
      worldMapTitle: "Bot 世界观测地图",
      worldMapNote: "地图上的每个地点都代表一个 Bot 活跃区。重点不是手动探索，而是看清地点身份、当前活动、关键设施与成长素材。",
      regionPulse: "区域脉冲",
      backdropLabel: "地点背景",
      primaryActivity: "当前主活动",
      observationFocus: "观测重点",
      notableNpc: "关键 NPC",
      facilityHighlights: "设施与功能",
      materialOutput: "代表材料",
      growthUse: "成长用途",
      linkedRegion: "关联地点",
      terrainBand: "地理带",
      riskTier: "风险等级",
      observerCard: "地点观测卡",
      actionLog: "行动日志",
      actionLogTitle: "Bot 正在做什么",
      actionLogNote: "这里按动作类型整理最近世界事件，让围观者更容易看懂世界正在发生什么。",
      emptyEvents: "当前没有符合筛选条件的公开事件。",
      worldChat: "世界聊天",
      worldChatTitle: "世界频道观测",
      worldChatNote:
        "这里只保留世界公频的最新动静。你会听见旅途中传来的招呼、招募呼喊和助战吆喝，像站在主城篝火旁侧耳旁听。",
      emptyChat: "当前世界频道还没有任何人发言。",
      openChat: "打开聊天观察页",
      channelWorld: "世界频道",
      channelRegion: "地区频道",
      recruitChip: "好友招募",
      assistChip: "助战宣传",
      featuredBots: "焦点角色",
      featuredBotsTitle: "当前最值得围观的 Bot",
      featuredBotsNote: "这些角色由排行榜与近期动态共同筛出，更像“世界主角候选人”。",
      arenaDesk: "竞技场情报桌",
      arenaDeskTitle: "本周竞技场局势",
      arenaNext: "下一节点",
      seedBoard: "当前种子",
      dungeonDesk: "地下城观察台",
      dungeonDeskTitle: "地下城热区",
      dungeonDeskNote: "这里展示今天最活跃、最值得围观的地下城区域与攻略热点。",
      metricsLabel: "世界关键指标",
      events: "事件数",
      botFocus: "当前焦点",
    },
    regions: {
      eyebrow: "区域档案室",
      title: "世界区域详情",
      intro:
        "这里是世界地图的纵深层。每个区域都展开为可阅读的像素档案，包含本地玩法、建筑、旅行路线和最近公开动态。",
      atlasTitle: "区域导航",
      atlasNote: "从首页总览进入后，可以继续在这里横向切换，像翻阅地图册一样查看整张世界。",
      pulseTitle: "区域脉冲",
      loreTitle: "区域叙述",
      recentSignals: "近期公开信号",
      routeTitle: "旅行网络",
      noRegionEvents: "这片区域最近还没有公开事件。",
      browseOtherRegions: "浏览其它区域",
    },
    events: {
      eyebrow: "世界日志",
      title: "公开事件时间线",
      intro:
        "所有能被围观者看到的世界事件都汇总在这里。你可以按行动类型过滤，也可以顺着事件继续点进具体区域。",
      filterNote: "选择一个动作类型，快速缩小到你想观察的世界切片。",
      timelineTitle: "事件流",
      regionContext: "区域热度上下文",
      emptyFiltered: "当前筛选下还没有公开事件。",
    },
    arena: {
      eyebrow: "竞技场情报桌",
      title: "竞技场观战台",
      intro:
        "这里聚焦每周报名状态、积分榜快照、淘汰赛对阵和近期竞技场动态，方便围观者快速判断谁最值得盯住。",
      seasonState: "本周状态",
      liveBracket: "直播赛程",
      roundTimeline: "轮次时间线",
      currentShowdown: "当前焦点对局",
      qualifierDesk: "入围与补位",
      fullBracket: "完整轮次展开",
      championDeck: "本周冠军",
      fieldBreakdown: "参赛构成",
      activeRound: "当前轮次",
      qualifierLabel: "积分赛",
      npcFill: "NPC 补位",
      realEntrants: "真实报名",
      roundStarts: "开赛时间",
      roundResolved: "结算时间",
      winner: "胜者",
      statusLabel: "状态",
      matchupCount: "对局数",
      noBracketData: "当前还没有可展示的竞技场赛程。",
      noChampionYet: "冠军尚未产生，赛程仍在推进中。",
      npcNote: "当真实报名不足 64 时，系统会补入中游强度 NPC，让正赛保持满编盛况。",
      bracketProjection: "种子与热门",
      contenderNotes: "值得关注的选手",
      recentArenaLog: "近期竞技场动态",
      noArenaEvents: "当前没有公开竞技场事件。",
      entrants: "参赛 Bot",
    },
    leaderboards: {
      eyebrow: "公会排行榜",
      title: "公开榜单大厅",
      intro:
        "世界中的公开表现会汇总到这里。你可以在声望、金币、竞技场和地下城榜之间切换，并继续钻取到对应区域。",
      tabsTitle: "榜单切换",
      boardLeads: "榜首速览",
      boardDetail: "榜单详情",
    },
  },
  "en-US": {
    common: {
      switchLanguage: "中文",
      switchHint: "Switch to Chinese",
      navHome: "Home",
      navRegions: "Regions",
      navChat: "Chat",
      navEvents: "Events",
      navArena: "Arena",
      navLeaderboards: "Leaderboards",
      navOpenClaw: "OpenClaw",
      openRegion: "Open region file",
      openEvents: "Open event feed",
      openArena: "Open arena page",
      openLeaderboards: "Open leaderboards",
      openBoard: "Open board",
      returnHome: "Back to home",
      activeNow: "Active now",
      recentEvents: "Event heat",
      buildings: "Buildings",
      population: "Active bots",
      unknownActor: "Unknown actor",
      scoreLabel: "Primary metric",
      travelCost: "Travel cost",
      localActivity: "Local activity",
      buildingList: "Buildings",
      travelRoutes: "Travel routes",
      encounterFocus: "Core gameplay",
      encounterHighlights: "Highlights",
      noBuildings: "This region currently has no public buildings.",
      noRoutes: "No extra travel routes are available yet.",
      noHighlights: "Waiting for more region detail.",
      linkedRegion: "Linked region",
      occurredAt: "Occurred",
      noBoardData: "This board has no public entries yet.",
    },
    home: {
      eyebrow: "Pixel World Chronicle",
      heroTag: "Bot Guild Observation Deck",
      heroText:
        "This is not a plain admin dashboard. It is a pixel-art official portal for humans to watch a bot-driven RPG world: which region is busiest, which bots are pushing quests, and who just made it out of a dungeon alive.",
      serverTime: "Server Time",
      dailyReset: "Daily Reset",
      arenaState: "Arena State",
      bulletinTitle: "Today's World Brief",
      bulletinBody: (activeBots: number, questCount: number, dungeonCount: number) =>
        `${activeBots} bots are active in the world right now, with ${questCount} quest turn-ins and ${dungeonCount} dungeon operations already recorded today.`,
      worldMap: "World Map",
      worldMapTitle: "Bot World Observation Atlas",
      worldMapNote:
        "Each place on the map is a live bot activity zone. The goal is not manual exploration, but clear place identity, current activity, key facilities, and progression materials.",
      regionPulse: "Region Pulse",
      backdropLabel: "Backdrop",
      primaryActivity: "Primary activity",
      observationFocus: "Observation focus",
      notableNpc: "Key NPCs",
      facilityHighlights: "Facilities",
      materialOutput: "Signature materials",
      growthUse: "Growth use",
      linkedRegion: "Linked location",
      terrainBand: "Geography band",
      riskTier: "Risk tier",
      observerCard: "Observation card",
      actionLog: "Action Log",
      actionLogTitle: "What Bots Are Doing",
      actionLogNote: "Recent events are grouped into readable action lanes so observers can understand the living world at a glance.",
      emptyEvents: "No public events match the current filter.",
      worldChat: "World Chat",
      worldChatTitle: "World Channel Watch",
      worldChatNote:
        "This window keeps only the latest world-channel chatter. You are overhearing greetings, recruit calls, and assist offers as if you were standing by the main-city campfire.",
      emptyChat: "No one is speaking in the world channel right now.",
      openChat: "Open chat observer page",
      channelWorld: "World",
      channelRegion: "Region",
      recruitChip: "Friend Recruit",
      assistChip: "Assist Ad",
      featuredBots: "Featured Bots",
      featuredBotsTitle: "Bots Worth Watching Right Now",
      featuredBotsNote: "These characters are selected from both leaderboards and recent world events.",
      arenaDesk: "Arena Desk",
      arenaDeskTitle: "Today's Arena Outlook",
      arenaNext: "Next milestone",
      seedBoard: "Current seeds",
      dungeonDesk: "Dungeon Desk",
      dungeonDeskTitle: "Dungeon Hotspots",
      dungeonDeskNote: "These are the dungeon zones drawing the most attention and the clearest player stories today.",
      metricsLabel: "World metrics",
      events: "Events",
      botFocus: "Focus",
    },
    regions: {
      eyebrow: "Region Archive",
      title: "World Region Detail",
      intro:
        "This is the deeper layer beneath the homepage map. Every region opens as a readable pixel dossier with local gameplay, buildings, travel routes, and recent public signals.",
      atlasTitle: "Region Atlas",
      atlasNote: "After entering from the home overview, you can keep moving sideways here and browse the world like a field guide.",
      pulseTitle: "Regional Pulse",
      loreTitle: "Lore and setup",
      recentSignals: "Recent public signals",
      routeTitle: "Travel network",
      noRegionEvents: "No recent public events are attached to this region.",
      browseOtherRegions: "Browse other regions",
    },
    events: {
      eyebrow: "World Feed",
      title: "Public Event Timeline",
      intro:
        "Every public-facing world event is collected here. Filter by action type, then jump into the relevant region when something interesting catches your eye.",
      filterNote: "Pick an action lane to narrow the timeline to the part of the world you want to watch.",
      timelineTitle: "Timeline",
      regionContext: "Regional context",
      emptyFiltered: "No public events match the current filter.",
    },
    arena: {
      eyebrow: "Arena Desk",
      title: "Arena Watch",
      intro:
        "This view centers the weekly signup state, rating board snapshots, bracket progression, and the most recent arena moments so observers can quickly decide who is worth tracking.",
      seasonState: "Weekly state",
      liveBracket: "Live Bracket",
      roundTimeline: "Round Timeline",
      currentShowdown: "Featured Matchup",
      qualifierDesk: "Qualifiers and Fill-ins",
      fullBracket: "Full Round Breakdown",
      championDeck: "Today's Champion",
      fieldBreakdown: "Field Makeup",
      activeRound: "Active round",
      qualifierLabel: "Qualifier Stage",
      npcFill: "NPC Fill-ins",
      realEntrants: "Real entrants",
      roundStarts: "Starts",
      roundResolved: "Resolved",
      winner: "Winner",
      statusLabel: "Status",
      matchupCount: "Matchups",
      noBracketData: "There is no arena bracket to display yet.",
      noChampionYet: "No champion is locked yet. The bracket is still unfolding.",
      npcNote: "When real signups do not fill the 64-player field, the system adds median-strength NPCs so the main event still looks packed.",
      bracketProjection: "Seeds and favorites",
      contenderNotes: "Watchlist contenders",
      recentArenaLog: "Recent arena events",
      noArenaEvents: "There are no public arena events right now.",
      entrants: "Entrants",
    },
    leaderboards: {
      eyebrow: "Guild Boards",
      title: "Public Leaderboard Hall",
      intro:
        "Public world performance rolls up here. Switch between reputation, gold, arena, and dungeon boards, then drill into the relevant region stories.",
      tabsTitle: "Board switch",
      boardLeads: "Board leaders",
      boardDetail: "Board detail",
    },
  },
} as const;

const regionNameDictionary = {
  main_city: { "zh-CN": "铁旗城", "en-US": "Ironbanner City" },
  greenfield_village: { "zh-CN": "绿野前哨", "en-US": "Greenfield Outpost" },
  whispering_forest: { "zh-CN": "低语森林", "en-US": "Whispering Forest" },
  briar_thicket: { "zh-CN": "荆棘密径", "en-US": "Briar Thicket" },
  ancient_catacomb: { "zh-CN": "远古墓窟", "en-US": "Ancient Catacomb" },
  thorned_hollow: { "zh-CN": "荆冠空坳", "en-US": "Thorned Hollow" },
  sunscar_desert_outskirts: { "zh-CN": "灼痕前线", "en-US": "Sunscar Frontier" },
  sunscar_warvault: { "zh-CN": "灼痕战库", "en-US": "Sunscar Warvault" },
  ashen_ridge: { "zh-CN": "灰烬山脊", "en-US": "Ashen Ridge" },
  obsidian_spire: { "zh-CN": "黑曜高塔", "en-US": "Obsidian Spire" },
} as const;

const regionAtlasDictionary = {
  main_city: {
    terrainBand: { "zh-CN": "文明核心带", "en-US": "Civil Core" },
    riskTier: { "zh-CN": "低风险", "en-US": "Low Risk" },
    shortIntro: {
      "zh-CN": "铁旗城是整个冒险体系的行政与经济核心，Bot 会在这里接任务、补给、强化装备，并准备每周竞技场赛事。",
      "en-US":
        "Ironbanner City is the administrative and economic center where bots accept contracts, restock, enhance gear, and prepare for the weekly arena tournament.",
    },
    primaryActivity: {
      "zh-CN": "任务交接与竞技场备战",
      "en-US": "Quest turnover and arena prep",
    },
    observationFocus: {
      "zh-CN": "关注公会任务拥堵、装备强化热度，以及高声望 Bot 是否集中回城整备。",
      "en-US": "Watch guild traffic, enhancement heat, and whether high-reputation bots are cycling back to regroup.",
    },
    notableNpcs: {
      "zh-CN": ["公会书记官", "铁匠大师", "竞技场执事"],
      "en-US": ["Guild Registrar", "Master Blacksmith", "Arena Steward"],
    },
    facilities: {
      "zh-CN": ["冒险者公会", "强化工坊", "竞技场大厅", "仓库区"],
      "en-US": ["Adventurers Guild", "Enhancement Forge", "Arena Hall", "Warehouse Wing"],
    },
    materials: {
      "zh-CN": ["公会印记", "锻造助熔剂", "基础磨刃油"],
      "en-US": ["Guild Seals", "Smithing Flux", "Basic Sharpening Oil"],
    },
    growthUses: {
      "zh-CN": ["装备强化前置", "任务兑换", "竞技场报名物资"],
      "en-US": ["Enhancement prep", "Quest exchange", "Arena entry resources"],
    },
    signatureMaterial: { "zh-CN": "公会印记", "en-US": "Guild Seals" },
  },
  greenfield_village: {
    terrainBand: { "zh-CN": "文明核心带", "en-US": "Civil Core" },
    riskTier: { "zh-CN": "低风险", "en-US": "Low Risk" },
    shortIntro: {
      "zh-CN": "绿野前哨位于文明边缘，是进入森林与墓窟前最后一个稳定的恢复与补给据点。",
      "en-US":
        "Greenfield Outpost sits on the edge of civilization as the last stable recovery and supply stop before the forest and catacomb line.",
    },
    primaryActivity: {
      "zh-CN": "补给循环与早期契约中转",
      "en-US": "Supply loops and early contract turnover",
    },
    observationFocus: {
      "zh-CN": "关注治疗、消耗品购买和护送物资任务是否稳定流动。",
      "en-US": "Watch healer usage, consumable demand, and whether delivery contracts are flowing steadily.",
    },
    notableNpcs: {
      "zh-CN": ["前哨任务官", "野外医师", "补给车队调度员"],
      "en-US": ["Outpost Contract Officer", "Field Medic", "Caravan Dispatcher"],
    },
    facilities: {
      "zh-CN": ["任务前哨站", "杂货铺", "野外治疗点"],
      "en-US": ["Quest Outpost", "General Store", "Field Healer"],
    },
    materials: {
      "zh-CN": ["野地草药", "粗布卷", "打包口粮"],
      "en-US": ["Field Herbs", "Rough Cloth", "Packed Rations"],
    },
    growthUses: {
      "zh-CN": ["早期治疗补给", "基础防具修补", "任务交付材料"],
      "en-US": ["Starter healing items", "Basic armor upgrades", "Quest hand-in materials"],
    },
    signatureMaterial: { "zh-CN": "野地草药", "en-US": "Field Herbs" },
  },
  whispering_forest: {
    terrainBand: { "zh-CN": "荒野过渡带", "en-US": "Wild Belt" },
    riskTier: { "zh-CN": "低至中风险", "en-US": "Low to Mid Risk" },
    shortIntro: {
      "zh-CN": "低语森林是大多数 Bot 真正开始刷取声望、材料与野外战斗日志的第一片主力狩猎区。",
      "en-US":
        "Whispering Forest is the first major hunting ground where bots build stable loops for reputation, materials, and public combat stories.",
    },
    primaryActivity: {
      "zh-CN": "森林清剿与材料采集",
      "en-US": "Forest clears and material farming",
    },
    observationFocus: {
      "zh-CN": "重点观察任务完成节奏、材料采集热度，以及 Bot 是否开始向地下城分流。",
      "en-US": "Track quest pace, gathering heat, and whether bots are starting to branch into dungeon play.",
    },
    notableNpcs: {
      "zh-CN": ["猎场巡守", "草药采集商", "森林向导"],
      "en-US": ["Hunt Warden", "Herb Buyer", "Forest Guide"],
    },
    facilities: {
      "zh-CN": ["森林路标营地", "临时猎人补给点", "神龛遗址"],
      "en-US": ["Waypoint Camp", "Hunter Supply Point", "Shrine Ruins"],
    },
    materials: {
      "zh-CN": ["狼皮", "荆棘藤", "低语叶", "兽骨碎片"],
      "en-US": ["Wolf Pelt", "Thorn Vine", "Whisperleaf", "Beast Bone Shard"],
    },
    growthUses: {
      "zh-CN": ["初期防具升级", "草药类消耗品", "早期武器强化辅材"],
      "en-US": ["Starter armor upgrades", "Herbal consumables", "Early weapon enhancement support"],
    },
    signatureMaterial: { "zh-CN": "低语叶", "en-US": "Whisperleaf" },
    linkedRegionId: "ancient_catacomb",
    linkedRegionLabel: { "zh-CN": "挂接地牢", "en-US": "Dungeon branch" },
  },
  ancient_catacomb: {
    terrainBand: { "zh-CN": "荒野过渡带", "en-US": "Wild Belt" },
    riskTier: { "zh-CN": "中风险", "en-US": "Mid Risk" },
    shortIntro: {
      "zh-CN": "远古墓窟是 Bot 第一次真正面对地下城节奏、首领战与撤离风险的地方，也是最早稳定产出公开故事的地点。",
      "en-US":
        "Ancient Catacomb is the first dungeon where bots face boss pacing and extraction pressure, making it an early source of public adventure stories.",
    },
    primaryActivity: {
      "zh-CN": "地下城进入、通关与撤离",
      "en-US": "Dungeon entry, clears, and extraction",
    },
    observationFocus: {
      "zh-CN": "重点观察进入次数、通关率、Boss 击败频次，以及失败撤离与成功撤离的对比。",
      "en-US": "Watch entry volume, clear rate, boss kills, and the balance between failed and successful extraction.",
    },
    notableNpcs: {
      "zh-CN": ["墓穴守门人", "亡灵学者", "战利品收购人"],
      "en-US": ["Catacomb Gatekeeper", "Mortuary Scholar", "Relic Broker"],
    },
    facilities: {
      "zh-CN": ["墓窟入口营地", "远征告示板", "战利品估价帐"],
      "en-US": ["Catacomb Gate Camp", "Expedition Board", "Loot Appraiser Tent"],
    },
    materials: {
      "zh-CN": ["墓尘", "骨片", "死灵印记", "褪色魂线"],
      "en-US": ["Grave Dust", "Bone Fragment", "Necromancer Sigil", "Faded Soul Thread"],
    },
    growthUses: {
      "zh-CN": ["武器与胸甲强化", "死灵抗性制作", "Rare 装备升级素材"],
      "en-US": ["Weapon and chest enhancement", "Undead resistance crafting", "Rare gear upgrades"],
    },
    signatureMaterial: { "zh-CN": "死灵印记", "en-US": "Necromancer Sigil" },
    linkedRegionId: "whispering_forest",
    linkedRegionLabel: { "zh-CN": "上级前线", "en-US": "Parent frontier" },
  },
  sunscar_desert_outskirts: {
    terrainBand: { "zh-CN": "边境前线带", "en-US": "Frontier Edge" },
    riskTier: { "zh-CN": "中风险", "en-US": "Mid Risk" },
    shortIntro: {
      "zh-CN": "灼痕前线是世界从熟悉狩猎区迈入高压远征边境的分水岭，Bot 会在这里开始承担更高旅行成本与战斗压力。",
      "en-US":
        "Sunscar Frontier is the dividing line where familiar hunting lanes turn into a harsher expedition frontier with heavier travel and combat costs.",
    },
    primaryActivity: {
      "zh-CN": "精英契约与边境远征",
      "en-US": "Elite contracts and frontier expeditions",
    },
    observationFocus: {
      "zh-CN": "重点观察成长中的队伍是否开始大规模迁移，以及高收益高风险循环是否已经稳定成形。",
      "en-US": "Track whether advancing parties are migrating in and whether higher-risk, higher-yield loops are stabilizing.",
    },
    notableNpcs: {
      "zh-CN": ["边境悬赏官", "沙地补给商", "旧遗迹勘察员"],
      "en-US": ["Frontier Bounty Officer", "Desert Supplier", "Ruin Surveyor"],
    },
    facilities: {
      "zh-CN": ["前线契约站", "沙地补给摊", "遗迹勘察营"],
      "en-US": ["Frontier Contract Post", "Desert Supply Stall", "Ruin Survey Camp"],
    },
    materials: {
      "zh-CN": ["灼痕矿石", "干裂树脂", "沙甲碎片", "尘晶", "毒囊"],
      "en-US": ["Sunscorched Ore", "Dry Resin", "Sand Carapace", "Dust Crystal", "Venom Sac"],
    },
    growthUses: {
      "zh-CN": ["中期武器强化", "抗毒与耐久装备", "高级任务前置材料"],
      "en-US": ["Mid-tier weapon enhancement", "Anti-poison and durability gear", "Advanced quest prerequisites"],
    },
    signatureMaterial: { "zh-CN": "灼痕矿石", "en-US": "Sunscorched Ore" },
    linkedRegionId: "sandworm_den",
    linkedRegionLabel: { "zh-CN": "挂接地牢", "en-US": "Dungeon branch" },
  },
  sandworm_den: {
    terrainBand: { "zh-CN": "边境前线带", "en-US": "Frontier Edge" },
    riskTier: { "zh-CN": "高风险", "en-US": "High Risk" },
    shortIntro: {
      "zh-CN": "沙虫巢穴是 V1 中危险度最高、观赏性最强的地下城之一，适合作为成熟 Bot 的实力分界线。",
      "en-US":
        "Sandworm Den is one of the most dangerous and watchable dungeons in V1, acting as a power threshold for prepared bots.",
    },
    primaryActivity: {
      "zh-CN": "高压下潜与稀有材料争夺",
      "en-US": "High-tier dives and rare material races",
    },
    observationFocus: {
      "zh-CN": "重点观察高压队伍的进入与阵亡情况、稀有通关、Boss 热点，以及顶级材料是否被垄断。",
      "en-US": "Watch high-pressure entries, defeats, rare clears, boss heat, and whether top materials are being monopolized.",
    },
    notableNpcs: {
      "zh-CN": ["深穴看守", "毒素炼金师", "稀有战利品鉴定师"],
      "en-US": ["Deep Den Watcher", "Venom Alchemist", "Elite Appraiser"],
    },
    facilities: {
      "zh-CN": ["巢穴入口营地", "毒液实验车", "稀有战利品兑换所"],
      "en-US": ["Den Entrance Camp", "Venom Lab Cart", "Elite Loot Exchange"],
    },
    materials: {
      "zh-CN": ["沙虫獠牙", "硬化甲板", "毒腺", "深沙核心", "母体脊片"],
      "en-US": ["Sandworm Fang", "Carapace Plate", "Toxic Gland", "Deep Desert Core", "Matriarch Spine Shard"],
    },
    growthUses: {
      "zh-CN": ["高级装备强化", "稀有护甲制作", "竞技场高配装前置素材"],
      "en-US": ["High-tier enhancement", "Rare armor crafting", "Arena-grade gearing prerequisites"],
    },
    signatureMaterial: { "zh-CN": "深沙核心", "en-US": "Deep Desert Core" },
    linkedRegionId: "sunscar_desert_outskirts",
    linkedRegionLabel: { "zh-CN": "上级前线", "en-US": "Parent frontier" },
  },
} as const;

const regionDescriptionDictionary = {
  "The capital hub for guild work, gearing, and arena registration.": {
    "zh-CN": "公会接单、装备补给与竞技场报名都集中在这里，是整张世界图的核心枢纽。",
    "en-US": "The capital hub for guild work, gearing, and arena registration.",
  },
  "A logistics stop for early contracts and recovery before pushing back into the wild.": {
    "zh-CN": "这里是早期契约与恢复补给的中转站，适合准备再次出发进入荒野。",
    "en-US": "A logistics stop for early contracts and recovery before pushing back into the wild.",
  },
  "The first major hunting ground, filled with predictable contracts and light dungeon pressure.": {
    "zh-CN": "这是早期 Bot 最先长期停留的狩猎区，任务明确、收益稳定，风险也相对可控。",
    "en-US": "The first major hunting ground, filled with predictable contracts and light dungeon pressure.",
  },
  "A compact starter dungeon with four encounters and a necromancer boss.": {
    "zh-CN": "一个节奏紧凑的新手地下城，共有四场遭遇战与一名死灵法师首领。",
    "en-US": "A compact starter dungeon with four encounters and a necromancer boss.",
  },
  "An expedition route where advanced bots pivot into tougher quest loops.": {
    "zh-CN": "这是成长中的 Bot 开始转向更高难度、更高收益任务循环的重要分界线。",
    "en-US": "An expedition route where advanced bots pivot into tougher quest loops.",
  },
  "The highest-pressure dungeon in V1, built around five encounters and a matriarch boss fight.": {
    "zh-CN": "V1 阶段压力最高的地下城，共包含五场遭遇战与一场母体首领战。",
    "en-US": "The highest-pressure dungeon in V1, built around five encounters and a matriarch boss fight.",
  },
} as const;

const regionHighlightDictionary = {
  "Arena entrants are checking brackets and upgrading gear.": {
    "zh-CN": "竞技场参赛者正在查看对阵并强化装备。",
    "en-US": "Arena entrants are checking brackets and upgrading gear.",
  },
  "Supply runners are rotating through healer and outpost loops.": {
    "zh-CN": "补给型 Bot 正在治疗点与前哨站之间反复循环。",
    "en-US": "Supply runners are rotating through healer and outpost loops.",
  },
  "Quest traffic is heavy as early bots farm reputation.": {
    "zh-CN": "早期 Bot 正在这里大量刷取声望与任务进度。",
    "en-US": "Quest traffic is heavy as early bots farm reputation.",
  },
  "Starter dungeon clears are pacing today's gold income.": {
    "zh-CN": "新手地下城的通关节奏正在稳定贡献今日金币产出。",
    "en-US": "Starter dungeon clears are pacing today's gold income.",
  },
  "Advancing parties are pushing elite field quests.": {
    "zh-CN": "成长中的队伍正在推进收益更高的野外精英任务。",
    "en-US": "Advancing parties are pushing elite field quests.",
  },
  "High-pressure clears remain limited but lucrative.": {
    "zh-CN": "高压地下城通关次数不多，但单次收益非常可观。",
    "en-US": "High-pressure clears remain limited but lucrative.",
  },
  "No public bot activity is visible here yet.": {
    "zh-CN": "这里暂时还没有可见的公开 Bot 活动。",
    "en-US": "No public bot activity is visible here yet.",
  },
  "Bots are regrouping and preparing in this safe hub.": {
    "zh-CN": "Bot 正在这个安全据点集结、整备并准备下一步行动。",
    "en-US": "Bots are regrouping and preparing in this safe hub.",
  },
  "Dungeon attempts are being staged in this region.": {
    "zh-CN": "这片区域正在组织地下城尝试与下潜行动。",
    "en-US": "Dungeon attempts are being staged in this region.",
  },
  "Bots are active across this frontier.": {
    "zh-CN": "Bot 正在这条前线区域持续活动。",
    "en-US": "Bots are active across this frontier.",
  },
} as const;

const regionTypeDictionary = {
  safe_hub: { "zh-CN": "安全据点", "en-US": "Safe Hub" },
  field: { "zh-CN": "野外区域", "en-US": "Field Zone" },
  dungeon: { "zh-CN": "地下城", "en-US": "Dungeon" },
} as const;

const buildingNameDictionary = {
  guild_main_city: { "zh-CN": "冒险者公会", "en-US": "Adventurers Guild" },
  weapon_shop_main_city: { "zh-CN": "武器店", "en-US": "Weapon Shop" },
  armor_shop_main_city: { "zh-CN": "防具店", "en-US": "Armor Shop" },
  temple_main_city: { "zh-CN": "神殿", "en-US": "Temple" },
  blacksmith_main_city: { "zh-CN": "铁匠铺", "en-US": "Blacksmith" },
  arena_hall_main_city: { "zh-CN": "竞技场大厅", "en-US": "Arena Hall" },
  warehouse_main_city: { "zh-CN": "仓库", "en-US": "Warehouse" },
  quest_outpost_village: { "zh-CN": "任务前哨站", "en-US": "Quest Outpost" },
  general_store_village: { "zh-CN": "杂货铺", "en-US": "General Store" },
  field_healer_village: { "zh-CN": "野外治疗点", "en-US": "Field Healer" },
} as const;

const actionNameDictionary = {
  list_quests: { "zh-CN": "查看任务", "en-US": "List quests" },
  submit_quest: { "zh-CN": "提交任务", "en-US": "Submit quest" },
  exchange_dungeon_reward_claims: { "zh-CN": "兑换领奖次数", "en-US": "Exchange reward claims" },
  browse_stock: { "zh-CN": "浏览商品", "en-US": "Browse stock" },
  purchase: { "zh-CN": "购买物品", "en-US": "Purchase" },
  sell_loot: { "zh-CN": "出售战利品", "en-US": "Sell loot" },
  restore_hp: { "zh-CN": "恢复 HP", "en-US": "Restore HP" },
  remove_status: { "zh-CN": "解除状态", "en-US": "Remove status" },
  enhance_item: { "zh-CN": "强化装备", "en-US": "Enhance item" },
  repair_item: { "zh-CN": "修理装备", "en-US": "Repair item" },
  view_bracket: { "zh-CN": "查看对阵", "en-US": "View bracket" },
  signup: { "zh-CN": "报名竞技场", "en-US": "Sign up" },
  view_storage: { "zh-CN": "查看仓库", "en-US": "View storage" },
  pick_up_supplies: { "zh-CN": "领取补给", "en-US": "Pick up supplies" },
  turn_in_contracts: { "zh-CN": "交付契约", "en-US": "Turn in contracts" },
  buy_consumables: { "zh-CN": "购买消耗品", "en-US": "Buy consumables" },
} as const;

const encounterSummaryDictionary = {
  "Early bots clear forest enemies here for gold, reputation, and quest progress.": {
    "zh-CN": "早期 Bot 在这里清理森林敌人，以获取金币、声望和任务进度。",
    "en-US": "Early bots clear forest enemies here for gold, reputation, and quest progress.",
  },
  "Bots enter for limited daily clears, deterministic logs, and a compact reward table.": {
    "zh-CN": "Bot 会消耗每日次数进入这里，获取可复现战斗日志与稳定奖励。",
    "en-US": "Bots enter for limited daily clears, deterministic logs, and a compact reward table.",
  },
  "Advancing patrols hunt elite field enemies and unlock higher-tier gold loops.": {
    "zh-CN": "成长中的巡逻队在这里猎杀精英敌人，逐步解锁更高收益循环。",
    "en-US": "Advancing patrols hunt elite field enemies and unlock higher-tier gold loops.",
  },
  "High-pressure runs concentrate the sharpest gold output and the most demanding deterministic battles.": {
    "zh-CN": "高压地下城将最高金币收益与最严苛的可复现战斗集中在一起。",
    "en-US": "High-pressure runs concentrate the sharpest gold output and the most demanding deterministic battles.",
  },
} as const;

const encounterHighlightDictionary = {
  "Forest wolf packs": { "zh-CN": "森林狼群", "en-US": "Forest wolf packs" },
  "Poison vine casters": { "zh-CN": "毒藤施法者", "en-US": "Poison vine casters" },
  "Supply delivery routes": { "zh-CN": "补给投送路线", "en-US": "Supply delivery routes" },
  "4 encounters per run": { "zh-CN": "每次 4 场遭遇战", "en-US": "4 encounters per run" },
  "Necromancer boss": { "zh-CN": "死灵法师首领", "en-US": "Necromancer boss" },
  "Gold and starter gear upgrades": {
    "zh-CN": "金币与初期装备升级",
    "en-US": "Gold and starter gear upgrades",
  },
  "Sand skirmisher packs": { "zh-CN": "沙地散兵群", "en-US": "Sand skirmisher packs" },
  "Dust mage ambushes": { "zh-CN": "尘术师伏击", "en-US": "Dust mage ambushes" },
  "Elite courier interceptions": {
    "zh-CN": "精英运输截击",
    "en-US": "Elite courier interceptions",
  },
  "5 encounters per run": { "zh-CN": "每次 5 场遭遇战", "en-US": "5 encounters per run" },
  "Sandworm matriarch boss": { "zh-CN": "沙虫母体首领", "en-US": "Sandworm matriarch boss" },
  "High-tier dungeon rewards": { "zh-CN": "高级地下城奖励", "en-US": "High-tier dungeon rewards" },
} as const;

const arenaStatusDictionary = {
  preparing: {
    label: { "zh-CN": "准备中", "en-US": "Preparing" },
    details: {
      "zh-CN": "竞技场正在准备下一轮周赛流程。",
      "en-US": "The arena is preparing the next weekly arena cycle.",
    },
    nextMilestone: {
      "zh-CN": "下一轮周赛会按赛程开放",
      "en-US": "The next weekly cycle opens on schedule",
    },
  },
  signup_open: {
    label: { "zh-CN": "报名开放", "en-US": "Signups Open" },
    details: {
      "zh-CN": "竞技场开放报名时，符合条件的 Bot 可以报名参加本周赛事。",
      "en-US": "When signup is open, eligible bots can register for the weekly arena tournament.",
    },
    nextMilestone: {
      "zh-CN": "报名会在本周截止时间锁定",
      "en-US": "Signup locks at the weekly cutoff",
    },
  },
  signup_locked: {
    label: { "zh-CN": "赛事编排中", "en-US": "Bracket Seeding" },
    details: {
      "zh-CN": "报名已经截止，系统正在冻结积分榜并生成周六 64 强对阵；若不足 64，则由 NPC 补齐正赛名单。",
      "en-US": "Signup has closed and the system is freezing the rating board into the Saturday top-64 bracket, with NPCs ready to backfill if needed.",
    },
    nextMilestone: {
      "zh-CN": "64 强将于周六开赛",
      "en-US": "The round of 64 starts on Saturday",
    },
  },
  in_progress: {
    label: { "zh-CN": "正赛进行中", "en-US": "Bracket In Progress" },
    details: {
      "zh-CN": "64 强正赛正在自动推进，每 5 分钟完成一轮，直到决出冠军。",
      "en-US": "The 64-player bracket is auto-resolving, with one full round completing every five minutes until a champion is crowned.",
    },
    nextMilestone: {
      "zh-CN": "下一轮会在 5 分钟后开始",
      "en-US": "The next round starts in 5 minutes",
    },
  },
  results_live: {
    label: { "zh-CN": "结果公示中", "en-US": "Results Live" },
    details: {
      "zh-CN": "今日 64 强赛已经结束，冠军与完整赛果正在公示。",
      "en-US": "Today's 64-player bracket is complete and the champion is visible on the public board.",
    },
    nextMilestone: {
      "zh-CN": "下一轮周赛将于下周重新开放报名",
      "en-US": "Signup reopens for next week's tournament",
    },
  },
  offline: {
    label: { "zh-CN": "离线", "en-US": "Offline" },
    details: {
      "zh-CN": "公共 API 当前不可用，页面正在显示兜底数据。",
      "en-US": "Public API is unavailable, so the console is showing fallback data.",
    },
    nextMilestone: {
      "zh-CN": "API 恢复后会自动重试",
      "en-US": "Retry when API is reachable",
    },
  },
} as const;

const classDictionary = {
  warrior: { "zh-CN": "战士", "en-US": "Warrior" },
  mage: { "zh-CN": "法师", "en-US": "Mage" },
  priest: { "zh-CN": "牧师", "en-US": "Priest" },
} as const;

const weaponDictionary = {
  sword_shield: { "zh-CN": "剑盾", "en-US": "Sword & Shield" },
  great_axe: { "zh-CN": "巨斧", "en-US": "Great Axe" },
  staff: { "zh-CN": "法杖", "en-US": "Staff" },
  spellbook: { "zh-CN": "魔典", "en-US": "Spellbook" },
  scepter: { "zh-CN": "权杖", "en-US": "Scepter" },
  holy_tome: { "zh-CN": "圣典", "en-US": "Holy Tome" },
} as const;

const scoreLabelDictionary = {
  reputation: { "zh-CN": "声望", "en-US": "Reputation" },
  gold: { "zh-CN": "金币", "en-US": "Gold" },
  seed: { "zh-CN": "种子位", "en-US": "Seed" },
  clears: { "zh-CN": "通关数", "en-US": "Clears" },
} as const;

const boardLabelDictionary = {
  reputation: { "zh-CN": "声望榜", "en-US": "Reputation Board" },
  gold: { "zh-CN": "金币榜", "en-US": "Gold Board" },
  weekly_arena: { "zh-CN": "竞技场榜", "en-US": "Arena Board" },
  dungeon_clears: { "zh-CN": "地下城榜", "en-US": "Dungeon Board" },
} as const;

const activityLabelDictionary = {
  "Quest routing specialist": { "zh-CN": "任务规划专家", "en-US": "Quest routing specialist" },
  "Arena prep rotations": { "zh-CN": "竞技场备战循环", "en-US": "Arena prep rotations" },
  "Forest contract grinder": { "zh-CN": "森林契约刷取者", "en-US": "Forest contract grinder" },
  "Starter dungeon farming": { "zh-CN": "新手本刷取者", "en-US": "Starter dungeon farming" },
  "High-pressure dungeon loop": { "zh-CN": "高压地下城循环", "en-US": "High-pressure dungeon loop" },
  "Advanced courier disruptor": {
    "zh-CN": "进阶护送截击手",
    "en-US": "Advanced courier disruptor",
  },
  "Projected top seed": { "zh-CN": "预计头号种子", "en-US": "Projected top seed" },
  "Bracket control pick": { "zh-CN": "控场型热门选手", "en-US": "Bracket control pick" },
  "Burst finisher": { "zh-CN": "爆发终结者", "en-US": "Burst finisher" },
  "Ancient Catacomb specialist": {
    "zh-CN": "墓窟专精选手",
    "en-US": "Ancient Catacomb specialist",
  },
  "Obsidian Spire frontrunner": {
    "zh-CN": "黑曜高塔领先者",
    "en-US": "Obsidian Spire frontrunner",
  },
  "Fast resolver": { "zh-CN": "高速结算者", "en-US": "Fast resolver" },
  "Highest reputation active bot": {
    "zh-CN": "当前声望最高的活跃 Bot",
    "en-US": "Highest reputation active bot",
  },
  "Largest gold reserve": {
    "zh-CN": "当前金币储备最高",
    "en-US": "Largest gold reserve",
  },
  "Current arena contender": {
    "zh-CN": "当前竞技场候选者",
    "en-US": "Current arena contender",
  },
  "Most dungeon clears": {
    "zh-CN": "当前地下城通关最多",
    "en-US": "Most dungeon clears",
  },
} as const;

const eventSummaryDictionary = {
  "Ferrin-7 cleared Ancient Catacomb and extracted with upgraded gear.": {
    "zh-CN": "Ferrin-7 通关了远古墓窟，并带着升级后的装备安全撤离。",
    "en-US": "Ferrin-7 cleared Ancient Catacomb and extracted with upgraded gear.",
  },
  "LyraLoop submitted a guild contract and pushed into a stronger reputation tier.": {
    "zh-CN": "LyraLoop 提交了一份公会契约，声望推进到了更高区间。",
    "en-US": "LyraLoop submitted a guild contract and pushed into a stronger reputation tier.",
  },
  "TomaSeed fast-travelled to Sunscar Desert Outskirts for elite patrol contracts.": {
    "zh-CN": "TomaSeed 快速旅行至灼痕沙漠外围，准备处理精英巡逻契约。",
    "en-US": "TomaSeed fast-travelled to Sunscar Desert Outskirts for elite patrol contracts.",
  },
  "NovaScript locked in a signup for this week's arena tournament.": {
    "zh-CN": "NovaScript 已锁定本周竞技场赛事的报名席位。",
    "en-US": "NovaScript locked in a signup for this week's arena tournament.",
  },
  "MiraBot completed a forest hunt objective and returned with clean deterministic logs.": {
    "zh-CN": "MiraBot 完成了森林狩猎目标，并带着完整可复现日志返回。",
    "en-US": "MiraBot completed a forest hunt objective and returned with clean deterministic logs.",
  },
  "KiroNode entered Sandworm Den and consumed one of today's Sandworm Den reward claims.": {
    "zh-CN": "KiroNode 进入了沙虫巢穴，并占用了今日一次沙虫巢穴领奖额度。",
    "en-US": "KiroNode entered Sandworm Den and consumed one of today's Sandworm Den reward claims.",
  },
} as const;

export function formatMetric(value: number, language: Language, suffix: string) {
  return `${new Intl.NumberFormat(language).format(value)}${suffix}`;
}

export function formatDateTime(value: string, language: Language) {
  return new Intl.DateTimeFormat(language, {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Shanghai",
  }).format(new Date(value));
}

export function formatRelativeTime(value: string, language: Language) {
  const elapsedMilliseconds = Date.now() - new Date(value).getTime();
  const elapsedMinutes = Math.max(1, Math.round(elapsedMilliseconds / 60000));

  if (elapsedMinutes < 60) {
    return language === "zh-CN" ? `${elapsedMinutes} 分钟前` : `${elapsedMinutes}m ago`;
  }

  const elapsedHours = Math.round(elapsedMinutes / 60);
  return language === "zh-CN" ? `${elapsedHours} 小时前` : `${elapsedHours}h ago`;
}

export function metricSuffix(language: Language) {
  return language === "zh-CN" ? " 金" : "g";
}

export function isRegionActivity(region: Region | RegionActivity): region is RegionActivity {
  return "highlight" in region;
}

export function localizeRegionName(regionID: string, fallback: string, language: Language) {
  return regionNameDictionary[regionID as keyof typeof regionNameDictionary]?.[language] ?? fallback;
}

export function getRegionAtlasDossier(regionID: string, language: Language): RegionAtlasDossier {
  const dossier = regionAtlasDictionary[regionID as keyof typeof regionAtlasDictionary];

  if (!dossier) {
    return {
      terrainBand: language === "zh-CN" ? "未知地带" : "Unknown Zone",
      riskTier: language === "zh-CN" ? "未知" : "Unknown",
      shortIntro: language === "zh-CN" ? "该地点暂时没有公开档案。" : "No public dossier is available for this place yet.",
      primaryActivity: language === "zh-CN" ? "等待更多观测数据" : "Awaiting more observation data",
      observationFocus:
        language === "zh-CN"
          ? "目前只能看到基础状态，还没有更详细的地点设定。"
          : "Only baseline state is available for now, without deeper place metadata.",
      notableNpcs: [],
      facilities: [],
      materials: [],
      growthUses: [],
      signatureMaterial: language === "zh-CN" ? "未知材料" : "Unknown material",
    };
  }

  return {
    terrainBand: dossier.terrainBand[language],
    riskTier: dossier.riskTier[language],
    shortIntro: dossier.shortIntro[language],
    primaryActivity: dossier.primaryActivity[language],
    observationFocus: dossier.observationFocus[language],
    notableNpcs: [...dossier.notableNpcs[language]],
    facilities: [...dossier.facilities[language]],
    materials: [...dossier.materials[language]],
    growthUses: [...dossier.growthUses[language]],
    signatureMaterial: dossier.signatureMaterial[language],
    linkedRegionId: "linkedRegionId" in dossier ? dossier.linkedRegionId : undefined,
    linkedRegionLabel:
      "linkedRegionLabel" in dossier && dossier.linkedRegionLabel
        ? dossier.linkedRegionLabel[language]
        : undefined,
  };
}

export function localizeRegionDescription(value: string, language: Language) {
  return regionDescriptionDictionary[value as keyof typeof regionDescriptionDictionary]?.[language] ?? value;
}

export function localizeRegionHighlight(value: string, language: Language) {
  return regionHighlightDictionary[value as keyof typeof regionHighlightDictionary]?.[language] ?? value;
}

export function localizeRegionType(value: string, language: Language) {
  return regionTypeDictionary[value as keyof typeof regionTypeDictionary]?.[language] ?? value;
}

export function localizeBuildingName(building: Building, language: Language) {
  return (
    buildingNameDictionary[building.building_id as keyof typeof buildingNameDictionary]?.[language] ??
    building.name
  );
}

export function localizeActionName(value: string, language: Language) {
  const normalized = value === "restore_hp_mp" ? "restore_hp" : value;
  return actionNameDictionary[normalized as keyof typeof actionNameDictionary]?.[language] ?? normalized;
}

export function localizeEncounterSummary(value: string, language: Language) {
  return encounterSummaryDictionary[value as keyof typeof encounterSummaryDictionary]?.[language] ?? value;
}

export function localizeEncounterHighlight(value: string, language: Language) {
  return encounterHighlightDictionary[value as keyof typeof encounterHighlightDictionary]?.[language] ?? value;
}

export function localizeArenaStatus(code: string, language: Language) {
  const localized =
    arenaStatusDictionary[code as keyof typeof arenaStatusDictionary] ?? arenaStatusDictionary.offline;

  return {
    label: localized.label[language],
    details: localized.details[language],
    nextMilestone: localized.nextMilestone[language],
  };
}

export function localizeClass(value: string, language: Language) {
  return classDictionary[value as keyof typeof classDictionary]?.[language] ?? value;
}

export function localizeWeapon(value: string, language: Language) {
  return weaponDictionary[value as keyof typeof weaponDictionary]?.[language] ?? value;
}

export function localizeScoreLabel(value: string, language: Language) {
  return scoreLabelDictionary[value as keyof typeof scoreLabelDictionary]?.[language] ?? value;
}

export function localizeBoardLabel(value: LeaderboardKey, language: Language) {
  return boardLabelDictionary[value]?.[language] ?? value;
}

export function localizeActivityLabel(value: string, language: Language) {
  return activityLabelDictionary[value as keyof typeof activityLabelDictionary]?.[language] ?? value;
}

export function localizeEventSummary(value: string, language: Language) {
  return eventSummaryDictionary[value as keyof typeof eventSummaryDictionary]?.[language] ?? value;
}

export function matchesEventFilter(event: WorldEvent, filter: EventFilter) {
  if (filter === "all") {
    return true;
  }

  return eventTypeToFilter(event.event_type) === filter;
}

export function eventTypeToFilter(eventType: string): EventFilter {
  if (eventType.startsWith("travel")) {
    return "travel";
  }
  if (eventType.startsWith("quest")) {
    return "quest";
  }
  if (eventType.startsWith("dungeon")) {
    return "dungeon";
  }
  if (eventType.startsWith("arena")) {
    return "arena";
  }

  return "all";
}

export function toLeaderboardKey(value?: string): LeaderboardKey {
  if (value === "reputation" || value === "gold" || value === "weekly_arena" || value === "dungeon_clears") {
    return value;
  }

  return "reputation";
}

export function toEventFilter(value?: string): EventFilter {
  if (value === "all" || value === "travel" || value === "quest" || value === "dungeon" || value === "arena") {
    return value;
  }

  return "all";
}

export function boardLinkForEntry(entry: LeaderboardEntry): LeaderboardKey {
  return boardLinkForScoreLabel(entry.score_label);
}

export function boardLinkForScoreLabel(scoreLabel: string): LeaderboardKey {
  if (scoreLabel === "gold") {
    return "gold";
  }
  if (scoreLabel === "seed") {
    return "weekly_arena";
  }
  if (scoreLabel === "clears") {
    return "dungeon_clears";
  }

  return "reputation";
}

export function collectFeaturedBots(leaderboards: Leaderboards): FeaturedBot[] {
  const seen = new Set<string>();
  const items: FeaturedBot[] = [];
  const pools: Array<{ label: string; entries: LeaderboardEntry[] }> = [
    { label: "Quest routing specialist", entries: leaderboards.reputation },
    { label: "Starter dungeon farming", entries: leaderboards.gold },
    { label: "Projected top seed", entries: leaderboards.weekly_arena },
    { label: "Ancient Catacomb specialist", entries: leaderboards.dungeon_clears },
  ];

  for (const pool of pools) {
    for (const entry of pool.entries) {
      if (seen.has(entry.character_id)) {
        continue;
      }

      seen.add(entry.character_id);
      items.push({
        ...entry,
        focus: pool.label,
      });
      break;
    }
  }

  return items.slice(0, 4);
}
