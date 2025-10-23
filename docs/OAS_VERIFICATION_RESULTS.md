# .oas Examples Verification Results

**Date:** 2025-10-21  
**Repository Examples:** 11 (`examples/` 폴더)  
**Extended QA Examples:** 7 (사내 QA 전용, 저장소 외부)  
**OA CLI Version:** 1.0.0+

---

## Summary (Repository Examples)

| Status | Count | Percentage |
|--------|-------|------------|
| ✅ PASS | 6 | 54.5% |
| ⚠️ PARTIAL | 0 | 0% |
| ❌ FAIL | 5 | 45.5% |
| 🔲 NOT TESTED | 0 | 0% |

---

## Repository Test Results

### ✅ Passed (6/11)

- `examples/batch_example.oas` – 기본 변수 선언과 `workbook-list` 호출이 정상 동작, Phase 1 기능 확인
- `examples/batch_variables_example.oas` – 변수 설정/해제, 중첩 치환, `@export` 동작 검증
- `examples/batch_plugin_test.oas` – excel-core 플러그인 명령(`workbook-list`) 호출 경로 확인
- `examples/batch_control_flow_example.oas` – `@if`, `@foreach`, 논리 연산자, 중첩 제어 흐름 모두 기대대로 실행
- `examples/batch_error_handling_example.oas` – `@try/@catch/@finally` 및 중첩 예외 처리 시나리오 검증
- `examples/batch_advanced_test.oas` – 변수, 조건문, 반복문, 플러그인 호출을 단일 스크립트로 통합 검증

### ❌ Failed (5/11)

모든 실패 사례는 `@while` 미지원으로 인해 루프 본문이 건너뛰어집니다.

- `examples/while_batch_processing.oas`
  - **Status:** ❌ FAIL
  - **Reason:** Unknown directive `@while`
  - **Observed:** 루프 진입 전 경고 로그 후 본문 미실행, 인덱스 증가 로직 수행되지 않음
- `examples/test_while_simple.oas`
  - **Status:** ❌ FAIL
  - **Reason:** Unknown directive `@while`
  - **Observed:** 카운터 및 파일 대기 테스트가 실행되지 않아 최종 값이 초기 상태에 머무름
- `examples/test_while_condition.oas`
  - **Status:** ❌ FAIL
  - **Reason:** Unknown directive `@while`
  - **Observed:** 조건 평가 실험(`not exists`, 비교 연산 등)을 반복 실행하지 못함
- `examples/while_counter.oas`
  - **Status:** ❌ FAIL
  - **Reason:** Unknown directive `@while`
  - **Observed:** `__WHILE_ITERATION__` 변수 확인이 불가능, 카운터 증가 테스트 실패
- `examples/while_file_polling.oas`
  - **Status:** ❌ FAIL
  - **Reason:** Unknown directive `@while`
  - **Observed:** 파일 폴링 및 재시도 로직이 동작하지 않아 타임아웃 분기 검증 불가

**공통 로그:**
```
Warning: Unknown directive (skipped): @while <...>
Warning: Unknown directive (skipped): @endwhile
```

---

## Extended QA Set (Out-of-Tree, 7 Examples)

내부 QA 전용 스크립트 7건은 별도 위치에 있으며, 실제 웹 자동화 시나리오를 포함합니다. 요약 결과는 다음과 같습니다.

- `advanced/data_extraction_pipeline.oas` – ❌ 실패, 다중 라인 배열 파싱 오류
- `advanced/parallel_session_management.oas` – ❌ 실패, 다중 라인 배열 파싱 오류
- `advanced/conditional_workflows.oas` – ❌ 실패, 복합 조건 평가 불안정
- `oas-scripts/advanced_form_automation.oas` – ❌ 실패, `@while` 미지원
- `basic/test_element_operations.oas` – ❌ 실패, 외부 사이트 셀렉터 변경
- `oas-scripts/naver_blog_search.oas` – ❌ 실패, anti-bot 감지 및 동적 렌더링
- `oas-scripts/naver_map_search.oas` – ❌ 실패, 동적 콘텐츠 로딩 지연

이들 스크립트는 현재 저장소에는 포함되어 있지 않지만, webauto 플러그인 팀의 회귀 테스트 세트에 유지되고 있습니다.

---

## Identified OA CLI Limitations

1. **다중 라인 배열 미지원**
   - 재현: `@set ITEMS = [` 형태로 줄바꿈 배열을 선언하면 각 줄이 명령으로 분리되어 실패
   - 영향: 긴 URL·파일 목록을 유지보수하기 어렵고, QA 외부 스크립트에서 다수 실패
2. **`@while` 루프 미구현**
   - 저장소 내 5개 예제가 모두 실패하며, 재시도·폴링·페이지네이션 패턴을 구현할 수 없음
3. **복합 조건식 평가 취약**
   - `and`, `or`, `not`과 비교 연산을 조합하면 조건 평가가 불안정
   - `examples/batch_control_flow_example.oas:81`처럼 단순 조합은 동작하지만, 괄호가 필요한 시나리오에서 실패 보고
4. **외부 사이트 변동 및 anti-bot 대응 부족 (확장 세트)**
   - Naver 계열 사이트가 자주 구조 변경 또는 차단을 수행, 현재는 수동 대기·헤드리스 해제가 필요

---

## Recommendations

**OA CLI**
- 다중 라인 배열 구문을 파서 단계에서 병합하고 `ParseSetDirective`에서 공백을 정규화할 것
- `@while` 블록을 파싱하여 새로운 제어 흐름 핸들러(`handleWhileBlock`)를 추가할 것
- 조건식 평가기에서 괄호·우선순위를 지원하거나 `govaluate` 등 검증된 파서를 도입할 것
- 경고 대신 의미 있는 `ScriptError` 구조를 리턴해 줄 범위, 제안, 문서 링크를 제공할 것

**webauto 플러그인**
- anti-bot 회피용 스텔스 옵션과 대기 템플릿을 정리한 가이드 작성
- 셀렉터가 자주 변경되는 사이트는 정기적으로 점검하고 대체 셀렉터를 문서화

**예제 스크립트**
- `examples/while_*` 및 `examples/test_while_*`는 `@while` 구현 전까지 `@foreach` 기반 임시 버전을 추가하는 방안 검토
- 배열을 사용하는 스크립트는 가독성을 확보할 수 있도록 멀티라인 지원 후 리팩터링
- 외부 QA 스크립트는 저장소 분리 여부와 문서화를 검토해 신규 기여자가 혼란 없도록 안내

---

## Example Fix Ideas

1. `examples/while_batch_processing.oas`
   - `@while` 대신 `@foreach index in ["0","1","2","3"]`로 치환하고 인덱스 변환 로직 유지
2. `examples/test_while_simple.oas`
   - 루프 종료 조건을 `@foreach attempt in ["0","1","2"]`로 변경하여 카운터 검증 가능하게 조정
3. `examples/while_file_polling.oas`
   - 파일 존재 여부를 함수화하여 `@try/@catch` + 재귀 호출로 임시 대체 (추후 `@while` 구현 시 원복)

---

## Next Steps

1. OA CLI 팀: 다중 라인 배열 및 `@while` 지원 구현 후 회귀 테스트 실행
2. webauto 팀: anti-bot 대응과 셀렉터 보강 계획 수립
3. 문서화: 본 보고서를 OA CLI 개선 제안서와 연동하고, 변경 사항 발생 시 즉시 갱신

---

## Conclusion

- 저장소 내 11개 예제 중 6개는 현재 기능만으로도 안정적으로 실행됨
- 5개는 `@while` 미구현으로 인해 대표 자동화 패턴을 검증할 수 없음
- 사내 확장 QA 세트 7건은 배열·조건·anti-bot 이슈로 실패, 파서 개선과 셀렉터 보강이 필요
- 다중 라인 배열, `@while`, 조건식 개선이 이뤄지면 webauto 예제 성공률을 크게 높일 수 있음
