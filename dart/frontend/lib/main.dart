// lib/main.dart
//
// Flutter demo: two dropdowns populated from a Go backend.
// - Fetches /states on startup and shows the first dropdown
// - Fetches /representatives?state=XX when a state is selected and populates the second dropdown
//
// Notes:
// - Add `http` package to pubspec.yaml: http: ^0.13.6 (or latest compatible)
// - Run on an emulator or device and make sure the backend is reachable.
//   If Flutter runs on Android emulator and your Go server is on host machine,
//   use "10.0.2.2" (Android emulator) instead of "localhost".
// - For iOS simulator "localhost" usually works.
// - This example focuses on clarity and comments rather than on perfect UX.

import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

// Replace with your backend base URL:
// - If running Go server on host and using Android emulator, use "http://10.0.2.2:8080"
// - If using iOS simulator or web, "http://localhost:8080" often works.
const String backendBaseUrl = "http://10.0.2.2:8080";

void main() {
  runApp(const MyApp());
}

// Basic data shapes matching the Go server responses:
class StateItem {
  final String code;
  final String name;
  StateItem({required this.code, required this.name});
  factory StateItem.fromJson(Map<String, dynamic> json) {
    return StateItem(
      code: json['code'] as String,
      name: json['name'] as String,
    );
  }
}

class Representative {
  final String id;
  final String firstName;
  final String lastName;
  final String party;
  final String district;
  Representative({
    required this.id,
    required this.firstName,
    required this.lastName,
    required this.party,
    required this.district,
  });
  factory Representative.fromJson(Map<String, dynamic> json) {
    return Representative(
      id: json['id'] as String,
      firstName: json['firstName'] as String,
      lastName: json['lastName'] as String,
      party: json['party'] as String,
      district: json['district'] as String,
    );
  }

  String get displayName =>
      '$firstName $lastName${party == '' ? ' (—)' : party}';
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Reps by State Demo',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: const StateRepPage(),
    );
  }
}

class StateRepPage extends StatefulWidget {
  const StateRepPage({super.key});
  @override
  State<StateRepPage> createState() => _StateRepPageState();
}

class _StateRepPageState extends State<StateRepPage> {
  // Lists for dropdowns
  List<StateItem> states = [];
  List<Representative> representatives = [];

  // Selected values
  String? selectedStateCode;
  String? selectedRepresentativeId;

  // Loading flags to show progress
  bool isLoadingStates = false;
  bool isLoadingReps = false;

  // Error message to display if network fails
  String? errorMessage;

  @override
  void initState() {
    super.initState();
    _loadStates();
  }

  // Fetch states from backend
  Future<void> _loadStates() async {
    setState(() {
      isLoadingStates = true;
      errorMessage = null;
    });

    try {
      final url = Uri.parse('$backendBaseUrl/states');
      final resp = await http.get(url).timeout(const Duration(seconds: 10));
      if (resp.statusCode != 200) {
        throw Exception('status ${resp.statusCode}');
      }
      final List<dynamic> data = json.decode(resp.body) as List<dynamic>;
      final loaded = data
          .map((e) => StateItem.fromJson(e as Map<String, dynamic>))
          .toList();
      setState(() {
        states = loaded;
        // Optionally preselect the first state
        if (states.isNotEmpty) {
          selectedStateCode = states.first.code;
          // Kick off loading of representatives for the default state
          _loadRepresentatives(states.first.code);
        }
      });
    } catch (e) {
      setState(() {
        errorMessage = 'Failed to load states: $e';
      });
    } finally {
      setState(() {
        isLoadingStates = false;
      });
    }
  }

  // Fetch representatives for a state
  Future<void> _loadRepresentatives(String stateCode) async {
    setState(() {
      isLoadingReps = true;
      errorMessage = null;
      representatives = [];
      selectedRepresentativeId = null;
    });

    try {
      final url = Uri.parse('$backendBaseUrl/representatives?state=$stateCode');
      final resp = await http.get(url).timeout(const Duration(seconds: 10));
      if (resp.statusCode != 200) {
        throw Exception('status ${resp.statusCode}');
      }
      final List<dynamic> data = json.decode(resp.body) as List<dynamic>;
      final loaded = data
          .map((e) => Representative.fromJson(e as Map<String, dynamic>))
          .toList();
      setState(() {
        representatives = loaded;
      });
    } catch (e) {
      setState(() {
        errorMessage = 'Failed to load representatives: $e';
      });
    } finally {
      setState(() {
        isLoadingReps = false;
      });
    }
  }

  // UI building
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Representatives by State')),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            // State dropdown area
            Row(
              children: [
                const Text('State:'),
                const SizedBox(width: 16),
                Expanded(
                  child: isLoadingStates
                      ? const LinearProgressIndicator()
                      : states.isEmpty
                      ? const Text('No states available')
                      : DropdownButton<String>(
                          isExpanded: true,
                          value: selectedStateCode,
                          items: states
                              .map(
                                (s) => DropdownMenuItem(
                                  value: s.code,
                                  child: Text('${s.name} (${s.code})'),
                                ),
                              )
                              .toList(),
                          onChanged: (val) {
                            if (val == null) return;
                            setState(() {
                              selectedStateCode = val;
                            });
                            // load representatives for the newly selected state
                            _loadRepresentatives(val);
                          },
                        ),
                ),
              ],
            ),
            const SizedBox(height: 24),

            // Representatives dropdown area
            Row(
              children: [
                const Text('Representative:'),
                const SizedBox(width: 16),
                Expanded(
                  child: isLoadingReps
                      ? const LinearProgressIndicator()
                      : representatives.isEmpty
                      ? const Text('No representatives')
                      : DropdownButton<String>(
                          isExpanded: true,
                          value: selectedRepresentativeId,
                          hint: const Text('Select a representative'),
                          items: representatives
                              .map(
                                (r) => DropdownMenuItem(
                                  value: r.id,
                                  child: Text(
                                    '${r.displayName} — ${r.district}',
                                  ),
                                ),
                              )
                              .toList(),
                          onChanged: (val) {
                            setState(() {
                              selectedRepresentativeId = val;
                            });
                          },
                        ),
                ),
              ],
            ),

            const SizedBox(height: 24),

            // Error display area
            if (errorMessage != null) ...[
              Text(errorMessage!, style: const TextStyle(color: Colors.red)),
            ],

            // Debugging helper: show selected values
            const SizedBox(height: 24),
            Text('Selected state: ${selectedStateCode ?? "-"}'),
            Text('Selected rep id: ${selectedRepresentativeId ?? "-"}'),
            const SizedBox(height: 30),

            Row(
              children: [
                const Text(
                  'By Zev Oster - https://zevosteressays.wordpress.com',
                  style: TextStyle(fontSize: 9, color: Colors.grey),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
