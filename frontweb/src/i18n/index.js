import { createI18n } from 'vue-i18n'

// 中文语言包
const zh = {
  common: {
    loading: '加载中...',
    success: '成功',
    error: '错误',
    warning: '警告',
    info: '信息',
    confirm: '确认',
    cancel: '取消',
    save: '保存',
    delete: '删除',
    edit: '编辑',
    view: '查看',
    back: '返回',
    refresh: '刷新',
    search: '搜索',
    more: '更多',
    yes: '是',
    no: '否',
    submit: '提交',
    reset: '重置',
    close: '关闭',
    open: '打开',
    language: '语言',
    chinese: '中文',
    english: '英文'
  },
  nav: {
    title: 'Go代码分析平台',
    home: '首页',
    staticAnalysis: '程序调用静态分析',
    runtimeAnalysis: '程序运行分析',
    about: '关于'
  },
  welcome: {
    title: 'Go代码分析平台',
    subtitle: '欢迎使用Go代码分析平台，一个强大的Go程序分析工具',
    about: {
      title: '关于本平台',
      content: 'Go代码分析平台是一个专为Go语言开发者设计的工具，旨在帮助您更好地理解和分析Go程序的运行机制和调用关系。通过可视化的方式展示程序的执行流程和函数调用关系，帮助您更高效地进行代码审查、性能优化和问题排查。'
    },
    features: '主要功能',
    runtime: {
      title: '程序运行分析',
      description: '通过收集程序运行时的调用信息，生成详细的函数调用链和执行路径，帮助您了解程序的实际执行流程。支持查看函数参数、返回值和执行耗时等详细信息。',
      instrumentation: {
        title: '项目插桩',
        placeholder: '请输入Go项目路径',
        button: '插桩',
        processing: '插桩中...',
        hint: '输入您的Go项目路径，点击"插桩"按钮将自动对项目进行插桩，以便进行运行时分析',
        pathError: '请输入项目路径',
        success: '项目插桩成功，现在可以运行您的程序进行分析',
        viewResults: '查看分析结果'
      },
      viewAnalysis: '查看运行分析'
    },
    static: {
      title: '程序调用静态分析',
      description: '通过静态分析Go程序的源代码，生成函数调用关系图，帮助您理解程序的结构和组件间的依赖关系。支持包级别和函数级别的调用关系分析，方便您进行代码重构和架构优化。',
      startAnalysis: '开始静态分析'
    },
    setlanguage: {
      title: '语言切换',
      description: '测试国际化功能是否正常工作。您可以在这里切换语言，查看界面元素是否正确翻译。',
      button: '语言切换'
    }
  },
  staticAnalysis: {
    title: '程序调用静态分析',
    tabs: {
      static: '静态调用分析',
      gitlab: 'GitLab改动影响分析'
    },
    form: {
      projectPath: '项目路径',
      projectPathPlaceholder: '请输入Go项目路径',
      startAnalysis: '开始分析',
      analyzing: '分析中...',
      tip: '提示',
      pathTip: '输入Go项目的完整路径，系统将自动分析项目的调用关系并生成数据库。',
      pathError: '请输入有效的项目路径',
      noTaskId: '服务器未返回有效的任务ID',
      analysisError: '启动分析失败'
    },
    options: {
      title: '分析选项',
      hideOptions: '隐藏选项',
      showOptions: '显示选项',
      optionsTip: '您可以根据需要配置以下分析选项，不同的选项会影响分析的精度和性能。',
      algorithm: '分析算法',
      cache: '缓存设置',
      enableCache: '启用缓存',
      cacheTip: '启用缓存可以加快后续分析速度，但可能占用更多磁盘空间',
      outputPath: '输出路径',
      outputPathPlaceholder: '输出文件路径',
      outputPathTip: '分析结果图像的输出路径，默认为 ./default.png',
      cachePath: '缓存路径',
      cachePathPlaceholder: '缓存文件路径',
      cachePathTip: '缓存文件的存储路径，默认为 ./cache.json',
      ignoreMethod: '忽略分析特定方法 (可选)',
      ignoreMethodPlaceholder: '输入包名或方法名, 如 pkg1,pkg2,pkg3',
      ignoreMethodTip: '忽略分析特定方法，留空则分析全部, 加速分析使用前缀匹配，如 github.com/gin-gonic/gin',
      algorithmTip: '选择用于构建调用图的算法，不同算法有不同的精度和性能特点',
      algorithms: {
        vta: 'VTA (变量类型分析)',
        rta: 'RTA (快速类型分析)',
        cha: 'CHA (类层次分析)',
        static: 'Static (静态分析)'
      }
    },
    monitor: {
      title: '分析任务状态监控',
      connected: '已连接',
      disconnected: '未连接',
      connecting: '连接中...',
      connectServer: '连接服务器',
      disconnect: '断开连接',
      taskId: '任务ID',
      status: {
        processing: '处理中',
        completed: '已完成',
        failed: '失败',
        not_found: '未找到'
      },
      logs: '分析日志',
      messages: '条消息',
      noMessages: '暂无消息',
      completed: '分析任务已完成，可以查看结果了。',
      refreshDb: '刷新数据库列表',
      failed: '分析任务失败',
      waitingTask: '等待任务启动',
      waitingTaskTip: '已连接到事件流，等待分析任务启动...',
      notConnected: '未连接到服务器',
      notConnectedTip: '点击"连接服务器"按钮建立事件流连接',
      parseError: '解析消息失败',
      tooManyParseErrors: '消息解析失败次数过多，请检查服务器响应格式',
      connectionError: '事件流连接失败，请检查网络或服务器状态',
      requestFailed: '请求失败',
      requestException: '请求异常'
    },
    guide: {
      title: '使用指南',
      welcome: '欢迎使用程序调用静态分析',
      description: '本工具可以帮助您分析Go程序的调用关系，生成静态调用图。',
      step1: '输入项目路径',
      step1Desc: '在上方输入框中输入您要分析的Go项目路径，点击"开始分析"按钮。',
      step2: '配置分析选项',
      step2Desc: '点击"显示选项"按钮，可以配置分析算法、缓存设置等高级选项。',
      step3: '等待分析完成',
      step3Desc: '系统将自动分析项目并生成数据库，分析完成后可以查看结果。',
      step4: '查看分析结果',
      step4Desc: '分析完成后，您可以选择数据库查看详细的分析结果。',
      refreshDb: '刷新数据库列表'
    },
    dbList: {
      title: '数据库文件列表',
      refresh: '刷新列表',
      fileName: '文件名',
      createTime: '创建时间',
      size: '大小',
      actions: '操作',
      viewAnalysis: '查看分析'
    },
    gitlab: {
      notAvailable: 'GitLab改动影响分析功能暂未开通',
      comingSoon: '该功能正在开发中，敬请期待...',
      backToStatic: '返回静态调用分析'
    }
  },
  runtimeAnalysis: {
    title: '程序运行分析',
    tabs: {
      runtimeAnalysis: '运行时分析大盘',
      functionAnalysis: '函数查询分析'
    },
    projectPath: {
      currentProjectPath: '当前项目路径',
      changeProjectPath: '更换项目',
      title: '运行时数据库文件',
      placeholder: '请输入Go项目路径',
      starting: "正在验证项目路径...",
      label: '文件路径',
      tip: '提示',
      verifying: '验证中...',
      startAnalysis: '开始分析',
      description: '输入Go项目的完整路径，系统将自动对项目进行插桩，以便进行运行时分析。'
    },
    instrumentation: {
      title: '项目插桩',
      placeholder: '请输入Go项目路径',
      startInstrumentation: '开始插桩',
      instrumenting: '插桩中...',
      tip: '提示',
      description: '输入Go项目的完整路径，系统将自动对项目进行插桩，以便进行运行时分析。'
    },
    statistics: {
      activeGoroutines: '活跃Goroutine',
      avgExecutionTime: '平均执行时间',
      maxCallDepth: '最大调用深度'
    },
    hotFunctions: {
      title: '热点函数分析',
      sortByCalls: '按调用次数',
      sortByTime: '按耗时',
      loading: '加载中...',
      loadingData: '正在加载热点函数数据...',
      noData: '暂无热点函数数据',
      functionName: '函数名',
      callCount: '调用次数',
      totalTime: '总耗时',
      avgTime: '平均耗时'
    },
    goroutineList: {
      title: 'Goroutine列表',
      currentPage: '当前页',
      gid: 'GID',
      initialFunction: '初始函数',
      callDepth: '调用深度',
      executionTime: '执行时间',
      actions: '操作',
      details: '详情',
      callGraph: '调用图',
      noData: '没有找到匹配的数据，请尝试其他搜索条件',
      prevPage: '上一页',
      nextPage: '下一页'
    },
    functionAnalysis:{
      title: '函数查询分析',
      search: '查询',
      inputFunctionName: '输入函数名称进行查询，支持模糊匹配',
      caller: '调用者',
      loading: '加载中...',
      analyzing: '正在分析函数调用关系...',
      noRelatedFunction: '未找到相关函数调用关系',
      tryOtherFunctionName: '请尝试其他函数名称',
      callRelationAnalysis: '调用关系分析',
      export: '导出',
      functionName: '函数名称',
      packagePath: '包路径',
      callLevel: '调用层级',
      callCount: '调用次数',
      avgTime: '平均耗时',
      operation: '操作'
    }
  }
}

// 英文语言包
const en = {
  common: {
    loading: 'Loading...',
    success: 'Success',
    error: 'Error',
    warning: 'Warning',
    info: 'Information',
    confirm: 'Confirm',
    cancel: 'Cancel',
    save: 'Save',
    delete: 'Delete',
    edit: 'Edit',
    view: 'View',
    back: 'Back',
    refresh: 'Refresh',
    search: 'Search',
    more: 'More',
    yes: 'Yes',
    no: 'No',
    submit: 'Submit',
    reset: 'Reset',
    close: 'Close',
    open: 'Open',
    language: 'Language',
    chinese: 'Chinese',
    english: 'English'
  },
  nav: {
    title: 'Go Code Analysis Platform',
    home: 'Home',
    staticAnalysis: 'Program Call Static Analysis',
    runtimeAnalysis: 'Program Runtime Analysis',
    about: 'About'
  },
  welcome: {
    title: 'Go Code Analysis Platform',
    subtitle: 'Welcome to Go Code Analysis Platform, a powerful tool for Go program analysis',
    about: {
      title: 'About This Platform',
      content: 'Go Code Analysis Platform is a tool designed specifically for Go language developers, aimed at helping you better understand and analyze the runtime mechanisms and call relationships of Go programs. It visualizes program execution flow and function call relationships, helping you perform code reviews, performance optimization, and troubleshooting more efficiently.'
    },
    features: 'Main Features',
    runtime: {
      title: 'Program Runtime Analysis',
      description: 'By collecting call information during program runtime, it generates detailed function call chains and execution paths, helping you understand the actual execution flow of the program. It supports viewing detailed information such as function parameters, return values, and execution time.',
      instrumentation: {
        title: 'Project Instrumentation',
        placeholder: 'Enter Go project path',
        button: 'Instrument',
        processing: 'Instrumenting...',
        hint: 'Enter your Go project path and click the "Instrument" button to automatically instrument the project for runtime analysis',
        pathError: 'Please enter a project path',
        success: 'Project instrumentation successful, you can now run your program for analysis',
        viewResults: 'View Analysis Results'
      },
      viewAnalysis: 'View Runtime Analysis'
    },
    static: {
      title: 'Program Call Static Analysis',
      description: 'By statically analyzing the source code of Go programs, it generates function call relationship diagrams, helping you understand the structure of the program and the dependencies between components. It supports call relationship analysis at both package and function levels, facilitating code refactoring and architecture optimization.',
      startAnalysis: 'Start Static Analysis'
    },
    setlanguage: {
      title: 'Language',
      description: 'Test whether the internationalization function is working properly. You can switch languages here to see if the interface elements are correctly translated.',
      button: 'Language Switching'
    }
  },
  staticAnalysis: {
    title: 'Program Call Static Analysis',
    tabs: {
      static: 'Static Call Analysis',
      gitlab: 'GitLab Change Impact Analysis'
    },
    form: {
      projectPath: 'Project Path',
      projectPathPlaceholder: 'Enter Go project path',
      startAnalysis: 'Start Analysis',
      analyzing: 'Analyzing...',
      tip: 'Tip',
      pathTip: 'Enter the full path of the Go project, and the system will automatically analyze the call relationships and generate a database.',
      pathError: 'Please enter a valid project path',
      noTaskId: 'Server did not return a valid task ID',
      analysisError: 'Failed to start analysis'
    },
    options: {
      title: 'Analysis Options',
      hideOptions: 'Hide Options',
      showOptions: 'Show Options',
      optionsTip: 'You can configure the following analysis options as needed. Different options will affect the accuracy and performance of the analysis.',
      algorithm: 'Analysis Algorithm',
      cache: 'Cache Settings',
      enableCache: 'Enable Cache',
      cacheTip: 'Enabling cache can speed up subsequent analysis, but may take up more disk space',
      outputPath: 'Output Path',
      outputPathPlaceholder: 'Output file path',
      outputPathTip: 'The output path of the analysis result image, default is ./default.png',
      cachePath: 'Cache Path',
      cachePathPlaceholder: 'Cache file path',
      cachePathTip: 'The storage path of the cache file, default is ./cache.json',
      ignoreMethod: 'ignore Analyze Specific Method (Optional)',
      ignoreMethodPlaceholder: 'Enter package name or method name',
      ignoreMethodTip: 'Ignore the analysis of specific methods, leave blank to analyze all, such as github.com/gin-gonic/gin',
      algorithmTip: 'Choose the algorithm for building the call graph, different algorithms have different accuracy and performance characteristics',
      algorithms: {
        vta: 'VTA (Variable Type Analysis)',
        rta: 'RTA (Rapid Type Analysis)',
        cha: 'CHA (Class Hierarchy Analysis)',
        static: 'Static (Static Analysis)'
      }
    },
    monitor: {
      title: 'Analysis Task Status Monitor',
      connected: 'Connected',
      disconnected: 'Disconnected',
      connecting: 'Connecting...',
      connectServer: 'Connect to Server',
      disconnect: 'Disconnect',
      taskId: 'Task ID',
      status: {
        processing: 'Processing',
        completed: 'Completed',
        failed: 'Failed',
        not_found: 'Not Found'
      },
      logs: 'Analysis Logs',
      messages: 'messages',
      noMessages: 'No messages',
      completed: 'Analysis task completed, you can view the results now.',
      refreshDb: 'Refresh Database List',
      failed: 'Analysis task failed',
      waitingTask: 'Waiting for task to start',
      waitingTaskTip: 'Connected to event stream, waiting for analysis task to start...',
      notConnected: 'Not connected to server',
      notConnectedTip: 'Click the "Connect to Server" button to establish an event stream connection',
      parseError: 'Failed to parse message',
      tooManyParseErrors: 'Too many message parsing failures, please check server response format',
      connectionError: 'Event stream connection failed, please check network or server status',
      requestFailed: 'Request failed',
      requestException: 'Request exception'
    },
    guide: {
      title: 'User Guide',
      welcome: 'Welcome to Program Call Static Analysis',
      description: 'This tool can help you analyze the call relationships of Go programs and generate static call graphs.',
      step1: 'Enter Project Path',
      step1Desc: 'Enter the path of the Go project you want to analyze in the input box above, and click the "Start Analysis" button.',
      step2: 'Configure Analysis Options',
      step2Desc: 'Click the "Show Options" button to configure advanced options such as analysis algorithm and cache settings.',
      step3: 'Wait for Analysis to Complete',
      step3Desc: 'The system will automatically analyze the project and generate a database. You can view the results after the analysis is completed.',
      step4: 'View Analysis Results',
      step4Desc: 'After the analysis is completed, you can select a database to view detailed analysis results.',
      refreshDb: 'Refresh Database List'
    },
    dbList: {
      title: 'Database File List',
      refresh: 'Refresh List',
      fileName: 'File Name',
      createTime: 'Create Time',
      size: 'Size',
      actions: 'Actions',
      viewAnalysis: 'View Analysis'
    },
    gitlab: {
      notAvailable: 'GitLab Change Impact Analysis feature is not available yet',
      comingSoon: 'This feature is under development, stay tuned...',
      backToStatic: 'Back to Static Call Analysis'
    }
  },
  runtimeAnalysis: {
    title: 'Program Runtime Analysis',
    tabs: {
      runtimeAnalysis: 'Runtime Analysis',
      functionAnalysis: 'Function Analysis'
    },
    projectPath: {
      currentProjectPath: 'Current Project Path',
      changeProjectPath: 'Change Project',
      title: 'Project Path & Instrumentation',
      placeholder: 'Enter Go project path',
      label: 'File Path',
      starting: "Verifying project path...",
      tip: 'Tip',
      startAnalysis: 'Start Analysis',
      verifying: 'Verifying...',
      description: 'Enter the full path of the Go project, and the system will automatically instrument the project for runtime analysis.'
    },
    instrumentation: {
      title: 'Project Instrumentation',
      placeholder: 'Enter Go project path',
      startInstrumentation: 'Start Instrumentation',
      instrumenting: 'Instrumenting...',
      tip: 'Tip',
      description: 'Enter the full path of the Go project, and the system will automatically instrument the project for runtime analysis.'
    },
    statistics: {
      activeGoroutines: 'Active Goroutines',
      avgExecutionTime: 'Avg Execution Time',
      maxCallDepth: 'Max Call Depth'
    },
    hotFunctions: {
      title: 'Hot Functions Analysis',
      sortByCalls: 'Sort by Calls',
      sortByTime: 'Sort by Time',
      loading: 'Loading...',
      loadingData: 'Loading hot functions data...',
      noData: 'No hot functions data available',
      functionName: 'Function Name',
      callCount: 'Call Count',
      totalTime: 'Total Time',
      avgTime: 'Avg Time'
    },
    goroutineList: {
      title: 'Goroutine List',
      currentPage: 'Current Page',
      gid: 'GID',
      initialFunction: 'Initial Function',
      callDepth: 'Call Depth',
      executionTime: 'Execution Time',
      actions: 'Actions',
      details: 'Details',
      callGraph: 'Call Graph',
      noData: 'No matching data found, please try other search criteria',
      prevPage: 'Previous',
      nextPage: 'Next'
    },
    functionAnalysis:{
      title: 'Function Analysis',
      search: 'Search',
      inputFunctionName: 'Input function name for query, supports fuzzy matching',
      caller: 'Caller',
      loading: 'Loading...',
      noRelatedFunction: 'No related function call relationship found',
      tryOtherFunctionName: 'Please try another function name',
      callRelationAnalysis: 'Call Relation Analysis',
      export: 'Export',
      functionName: 'Function Name',
      packagePath: 'Package Path',
      callLevel: 'Call Level',
      callCount: 'Call Count',
      avgTime: 'Avg Time',
      operation: 'Operation'
    }
  }
}

// 创建i18n实例
const i18n = createI18n({
  legacy: false, // 使用组合式API
  globalInjection: true, // 全局注入 $t 函数
  locale: localStorage.getItem('locale') || 'zh', // 默认语言
  fallbackLocale: 'zh', // 回退语言
  silentTranslationWarn: true, // 静默翻译警告
  messages: {
    zh,
    en
  }
})

export default i18n 